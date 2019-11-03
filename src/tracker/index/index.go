package index

import (
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"tracker/captcha"
	"tracker/config"
	"tracker/db"
	"tracker/log"
	"tracker/metainfo"
	"tracker/model"
	"tracker/scrape"
	"tracker/util"
)

type Server struct {
	cfg        *config.IndexConfig
	mux        *http.ServeMux
	tmpl       *template.Template
	DB         db.DB
	Cfg_scrape *config.ScrapeConfig
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
func (s *Server) Error(w http.ResponseWriter, msg string, j bool) {
	var err error
	p := map[string]interface{}{
		"Error": msg,
	}
	if j {
		w.Header().Set("Content-Type", "text/json")
		err = json.NewEncoder(w).Encode(p)
	} else {
		err = s.tmpl.ExecuteTemplate(w, "error.html.tmpl", p)
	}
	if err != nil {
		log.Errorf("error rendering error page: %s", msg)
	}
}

// filter torrent file to have the members we want
func (s *Server) filterTorrent(t *metainfo.TorrentFile) *metainfo.TorrentFile {
	urls := t.GetAllAnnounceURLS()
	us := s.cfg.TrackerURL.String()
	// add our tracker url to file if it's not there
	found := false
	for _, u := range urls {
		if us == u {
			found = true
			break
		}
	}

	// add if not found
	if !found {
		t.Announce = us
		var alist [][]string
		for _, a := range urls {
			alist = append(alist, []string{a})
		}
		t.AnnounceList = alist
		urls = append(urls, us)
	}
	return t
}

func (s *Server) shouldAUTH(r *http.Request) bool {
	return r.URL.Query().Get("auth") == "1"
}

func (s *Server) shouldJSON(r *http.Request) bool {
	return r.URL.Query().Get("t") == "json"
}

func (s *Server) shouldATOM(r *http.Request) bool {
	return r.URL.Query().Get("t") == "atom" || strings.HasSuffix(r.URL.Path, ".atom.xml")
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	j := s.shouldJSON(r)
	feed := s.shouldATOM(r)
	if r.Method == "GET" {
		tags, err := s.DB.ListPopularTags(30)
		if err != nil {
			s.Error(w, err.Error(), j)
			return
		}
		var torrents []model.Torrent
		var selectedTag *model.Tag
		name := r.URL.Query().Get("q")
		tag := r.URL.Query().Get("id")

		if name != "" {
			selectedTag, err = s.DB.GetTagByName(name)
		}

		if selectedTag == nil && tag != "" {
			id, err := strconv.Atoi(tag)
			if err != nil {
				s.Error(w, err.Error(), j)
				return
			}
			if id > 0 {
				selectedTag, err = s.DB.GetTagByID(uint64(id))
				if err != nil {
					s.Error(w, err.Error(), j)
					return
				}
			}
		}

		if selectedTag != nil {
			torrents, err = s.DB.FindTorrentsWithTag(*selectedTag)
		}

		if feed && selectedTag != nil {
			f := &model.AtomFeed{
				Title:   selectedTag.Name,
				ID:      fmt.Sprintf("torrents-tag-%d", selectedTag.ID),
				BaseURL: r.URL,
				Domain:  r.Host,
			}
			for _, torrent := range torrents {
				torrent.Domain = r.Host
				f.Entries = append(f.Entries, torrent)
			}
			w.Header().Set("Content-Type", "application/atom+xml")
			enc := xml.NewEncoder(w)
			err = enc.Encode(f)
		} else if j {
			var jtorrents []model.Torrent
			for _, torrent := range torrents {
				torrent.Domain = r.Host
				jtorrents = append(jtorrents, torrent)
			}
			err = json.NewEncoder(w).Encode(map[string]interface{}{
				"Torrents": jtorrents,
			})
		} else if feed {
			f := &model.AtomFeed{
				Title:   "popular tags",
				ID:      "torrent-popular-tags",
				BaseURL: r.URL,
				Domain:  r.Host,
			}
			for _, tag := range tags {
				tag.Domain = r.Host
				f.Entries = append(f.Entries, tag)
			}
			w.Header().Set("Content-Type", "application/atom+xml")
			enc := xml.NewEncoder(w)
			err = enc.Encode(f)

		} else {
			results := ""
			if len(torrents) == 1 {
				results = "result"
			} else {
				results = "results"
			}
			u := r.URL
			q := u.Query()
			q.Add("t", "atom")
			u.RawQuery = q.Encode()
			u.Host = r.Host
			u.Scheme = "http"
			err = s.tmpl.ExecuteTemplate(w, "search.html.tmpl", map[string]interface{}{
				"PopularTags":  tags,
				"Site":         s.cfg.SiteName,
				"Torrents":     torrents,
				"SelectedTag":  selectedTag,
				"Search":       tag != "" || name != "",
				"SearchTag":    name,
				"FeedURL":      u.String(),
				"ResultString": results, // TODO
			})
		}
		if err != nil {
			log.Errorf("Failed to render search page: %s", err)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) addTorrent(w http.ResponseWriter, r *http.Request, cat model.Category) {
	j := s.shouldJSON(r)
	a := s.shouldAUTH(r)
	store := s.DB
	if store == nil {
		s.Error(w, "internal error: no storage backend", j)
		return
	}
	defer r.Body.Close()
	tags := r.FormValue("torrent-tags")
	name := r.FormValue("torrent-name")
	description := strings.TrimFunc(r.FormValue("torrent-description"), util.IsSpace)

	s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if len(description) == 0 {
			s.Error(w, "no description", j)
			return
		}

		t := new(metainfo.TorrentFile)
		f, h, err := r.FormFile("torrent-file")
		if name == "" {
			name = h.Filename
		}
		if name == "" {
			s.Error(w, "torrent name not specified", j)
			return
		}
		if err != nil {
			s.Error(w, err.Error(), j)
			return
		}
		err = t.Decode(f)
		if err != nil {
			s.Error(w, err.Error(), j)
			return
		}

		if t.Info.Private > 0 {
			s.Error(w, "private torrents not allowed", j)
			return
		}

		t = s.filterTorrent(t)

		torrent, err := store.FindTorrentByInfohash(t.Infohash())
		if err != nil {
			s.Error(w, err.Error(), j)
			return
		}
		if torrent != nil {
			s.Error(w, "duplicate torrent", j)
			return
		}

		torrent = &model.Torrent{
			Uploaded:     time.Now().Unix(),
			PieceSize:    t.Info.PieceLength,
			IH:           t.Infohash(),
			Size:         t.TotalSize(),
			AnnounceURLS: t.GetAllAnnounceURLS(),
			Name:         name,
			Category:     cat,
		}

		// set tags
		tags = strings.ToLower(tags)

		for _, tag := range strings.Split(tags, ",") {
			for _, tag1 := range strings.Split(tags, ",") {
				if tag == tag1 {
					s.Error(w, "tags error", j)
					return
				}
			}
		}
		for _, tag := range strings.Split(tags, ",") {
			tname := strings.Replace(strings.Trim(tag, " "), " ", "-", -1)
			torrent.Tags = append(torrent.Tags, model.Tag{
				Name: tname,
			})
		}
		err = store.StoreTorrent(torrent, t)
		if err != nil {
			s.Error(w, "could not store torrent: "+err.Error(), j)
			return
		}
		// store torrent seed file
		fpath := filepath.Join(s.cfg.TorrentsDir, fmt.Sprintf("%s.torrent", torrent.InfoHash()))
		var file *os.File
		file, err = os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			s.Error(w, "could not open file: "+err.Error(), j)
			return
		}
		err = t.Encode(file)
		file.Close()
		if err != nil {
			s.Error(w, err.Error(), j)
			return
		}

		// TODO: don't ignore
		s.DB.InsertComment(description, torrent.IH)

		// success
		if j {
			u := &url.URL{
				Scheme: "http",
				Host:   r.Host,
				Path:   torrent.DownloadLink(),
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"Error":    nil,
				"InfoHash": torrent.InfoHash(),
				"URL":      u.String(),
			})
		} else {

			location := torrent.PageLocation()
			if a {
				location += "?auth=1"
			}
			http.Redirect(w, r, location, http.StatusFound)

		}
	}, w, r)
}

func (s *Server) NotFound(w http.ResponseWriter, p map[string]interface{}, j bool) {
	w.WriteHeader(http.StatusNotFound)
	if j {
		json.NewEncoder(w).Encode(map[string]string{
			"Error": "not found",
		})
	} else {
		s.tmpl.ExecuteTemplate(w, "not-found.html.tmpl", p)
	}
}

func (s *Server) handleCategoryPage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	j := s.shouldJSON(r)
	a := s.shouldAUTH(r)
	feed := s.shouldATOM(r)

	p := map[string]interface{}{
		"Message": "No Such Category",
	}
	catid, err := strconv.Atoi(strings.Trim(path[3:], "/"))
	if err != nil {

		s.NotFound(w, p, j)

		return
	}
	cat, err := s.DB.GetCategoryByID(catid)
	if err != nil {
		s.Error(w, err.Error(), j)
		return
	}
	if cat == nil {
		s.NotFound(w, p, j)
		return
	}

	if r.Method == "GET" {

		pagestr := r.URL.Query().Get("p")

		if pagestr == "" {
			pagestr = "0"
		}
		page, err := strconv.Atoi(pagestr)
		if err != nil {
			s.Error(w, err.Error(), j)
		}

		perpage := 50

		torrents, err := s.DB.FindTorrentsInCategory(cat, perpage, perpage*page)
		if err != nil {
			s.Error(w, err.Error(), j)
			return
		}
		isEmpty := len(torrents) == 0
		if feed {
			w.Header().Set("Content-Type", "application/atom+xml")
			if isEmpty {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			f := &model.AtomFeed{
				Title:   cat.Name,
				ID:      fmt.Sprintf("torrents-category-%d", cat.ID),
				BaseURL: r.URL,
				Domain:  r.Host,
			}
			for _, torrent := range torrents {
				torrent.Domain = r.Host
				f.Entries = append(f.Entries, torrent)
			}
			enc := xml.NewEncoder(w)
			err = enc.Encode(f)
		} else if j {
			var jtorrents []model.Torrent
			for _, torrent := range torrents {
				torrent.Domain = r.Host
				torrent.Category = *cat
				jtorrents = append(jtorrents, torrent)
			}
			err = json.NewEncoder(w).Encode(map[string]interface{}{
				"Torrents": jtorrents,
				"Category": cat,
			})
		} else {
			var nextPage, prevPage int
			if page > 0 {
				prevPage = page - 1
			}
			nextPage = page + 1

			p := map[string]interface{}{
				"TrackerURL":  s.cfg.TrackerURL.Host,
				"Domain":      r.Host,
				"Torrents":    torrents,
				"Category":    cat,
				"Site":        s.cfg.SiteName,
				"NextPage":    nextPage,
				"PrevPage":    prevPage,
				"HasNextPage": !isEmpty,
				"HasPrevPage": page > 0,
			}

			var ok bool
			ok, _, err = s.checkAuth(r)
			if ok {

			} else {
				if a {
					w.Header().Set("WWW-Authenticate", "Basic")
					w.WriteHeader(http.StatusUnauthorized)
					return
				} else {
					cid := captcha.New()
					p["Captcha"] = cid
				}
			}
			if err == nil {
				err = s.tmpl.ExecuteTemplate(w, "category.html.tmpl", p)
			} else {
				s.Error(w, err.Error(), j)
			}
		}

		if err != nil {
			log.Errorf("Failed to render category page: %s", err)
		}
	} else if r.Method == "POST" {
		s.addTorrent(w, r, *cat)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) serveTorrent(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path[3:], "/")
	if strings.Count(path, "..") > 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if strings.HasSuffix(path, ".torrent") {
		http.ServeFile(w, r, filepath.Join(s.cfg.TorrentsDir, path))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) serveFrontPage(w http.ResponseWriter, r *http.Request) {
	j := s.shouldJSON(r)
	feed := s.shouldATOM(r)
	path := r.URL.Path
	if path != "/" && path != "/index.html" {
		http.NotFound(w, r)
		return
	}
	cats, err := s.DB.GetAllCategories()
	if err != nil {
		s.Error(w, "failed to fetch categories: "+err.Error(), j)
		return
	}
	torrents, err := s.DB.GetFrontPageTorrents()
	if err != nil {
		s.Error(w, "failed to fetch front page torrents: "+err.Error(), j)
		return
	}

	if feed {
		u := r.URL
		u.Host = r.Host
		u.Scheme = "http"
		f := &model.AtomFeed{
			Title:   "Recent Uploads",
			ID:      "torrent-recent-uploads",
			BaseURL: u,
			Domain:  r.Host,
		}
		for _, torrent := range torrents {
			torrent.Domain = r.Host
			f.Entries = append(f.Entries, torrent)
		}
		w.Header().Set("Content-Type", "application/atom+xml")
		enc := xml.NewEncoder(w)
		err = enc.Encode(f)
	} else if j {
		var jtorrents []model.Torrent
		for _, torrent := range torrents {
			torrent.Domain = r.Host
			jtorrents = append(jtorrents, torrent)
		}
		err = json.NewEncoder(w).Encode(map[string]interface{}{
			"Categories": cats,
			"Torrents":   jtorrents,
		})
	} else {
		err = s.tmpl.ExecuteTemplate(w, "frontpage.html.tmpl", map[string]interface{}{
			"Domain":     r.Host,
			"Categories": cats,
			"Torrents":   torrents,
			"Site":       s.cfg.SiteName,
		})
	}
	if err != nil {
		log.Errorf("Failed to render front page: %s", err)
	}
}

func (s *Server) serveTorrentInfo(w http.ResponseWriter, r *http.Request) {
	j := s.shouldJSON(r)
	feed := s.shouldATOM(r)
	a := s.shouldAUTH(r)
	ihstr := strings.Trim(r.URL.Path[3:], "/")
	ihbytes, err := hex.DecodeString(ihstr)
	if err == nil {
		if len(ihbytes) == 20 {
			var ih [20]byte
			copy(ih[:], ihbytes)
			var t *model.Torrent
			t, err = s.DB.FindTorrentByInfohash(ih)
			if r.Method == "GET" {
				if t != nil {
					// get torrent files
					files, err := s.DB.GetTorrentFiles(ih)
					if err != nil {
						s.Error(w, err.Error(), j)
						return
					}
					// get comments
					comments, err := s.DB.GetCommentsForTorrent(t)
					if err != nil {
						s.Error(w, err.Error(), j)
						return
					}
					// get tags
					tags, err := s.DB.GetTorrentTags(t)
					if err != nil {
						s.Error(w, err.Error(), j)
						return
					}

					// get captcha
					t.Domain = r.Host

					_, sm := scrape.GetScrapeByInfoHash(s.Cfg_scrape.File_path, s.Cfg_scrape.URL, string( hex.EncodeToString(t.IH[:]) ))
					if len(sm) == 0	{
						item:=scrape.Files{
							Downloaded: 0,
							Complete: 0,
							Incomplete: 0,
						}
						sm=append(sm, item)
					}
					fmt.Println( sm )

					p := map[string]interface{}{
						"Tags":       tags,
						"Torrent":    t,
						"Files":      files,
						"Site":       s.cfg.SiteName,
						"Comments":   comments,
						"Domain":     r.Host,
						"Downloaded": sm[0].Downloaded,
						"Complete":   sm[0].Complete,
						"Incomplete": sm[0].Incomplete,
					}

					var ok bool
					ok, _, err = s.checkAuth(r)
					if ok {
					} else {
						if a {
							w.Header().Set("WWW-Authenticate", "Basic")
							w.WriteHeader(http.StatusUnauthorized)
							return
						} else {
							p["Captcha"] = captcha.New()
						}
					}
					if feed {
						u := r.URL
						u.Host = r.Host
						u.Scheme = "http"
						f := &model.AtomFeed{
							Title:   fmt.Sprintf("Comments on %s", t.Name),
							ID:      fmt.Sprintf("comments-%s", ihstr),
							BaseURL: u,
							Domain:  r.Host,
						}
						for _, comment := range comments {
							comment.Domain = r.Host
							comment.Torrent = t
							f.Entries = append(f.Entries, comment)
						}
						w.Header().Set("Content-Type", "application/atom+xml")
						enc := xml.NewEncoder(w)
						err = enc.Encode(f)
					} else if j {
						err = json.NewEncoder(w).Encode(p)
					} else {
						err = s.tmpl.ExecuteTemplate(w, "torrent.html.tmpl", p)
					}
					if err != nil {
						log.Errorf("Failed to render torrent page: %s", err)
					}
					return
				}
			} else if r.Method == "POST" {
				s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
					action := strings.TrimFunc(r.FormValue("action"), util.IsSpace)
					if action == "comment" {
						comment := strings.TrimFunc(r.FormValue("comment"), util.IsSpace)
						if len(comment) > 0 {
							err = s.DB.InsertComment(comment, t.IH)
							if err == nil {
								location := t.PageLocation()
								if a {
									location += "?auth=1"
								}
								http.Redirect(w, r, location, http.StatusFound)
							} else {
								s.Error(w, err.Error(), j)
							}
						} else {
							s.Error(w, "Empty comment", j)
						}
					} else if action == "tag" {
						addTags := strings.Split(r.FormValue("add"), ",")
						for idx, tag := range addTags {
							addTags[idx] = strings.TrimFunc(tag, util.IsSpace)
						}

						delTags := strings.Split(r.FormValue("del"), ",")
						for idx, tag := range delTags {
							delTags[idx] = strings.TrimFunc(tag, util.IsSpace)
						}

						if len(addTags)+len(delTags) == 0 {
							s.Error(w, "No tags changed", j)
							return
						}

						existing, err := s.DB.GetTorrentTags(t)
						if err != nil {
							s.Error(w, err.Error(), j)
							return
						}
						var addTagsTorrent []string
						for _, tag := range addTags {
							if len(tag) > 0 {
								tag = strings.ToLower(tag)
								found := false
								for _, val := range existing {
									if strings.ToLower(val.Name) == tag {
										found = true
										break
									}
								}
								if !found {
									addTagsTorrent = append(addTagsTorrent, tag)
								}
							}
						}
						if len(addTags) > 0 {
							tags, err := s.DB.EnsureTags(addTagsTorrent)
							if err == nil {
								err = s.DB.AddTorrentTags(tags, t)
							}
							if err != nil {
								s.Error(w, err.Error(), j)
								return
							}
						}
						if len(delTags) > 0 {
							tags, err := s.DB.EnsureTags(delTags)
							if err == nil {
								err = s.DB.DelTorrentTags(tags, t)
							}
							if err != nil {
								s.Error(w, err.Error(), j)
								return
							}
						}
						http.Redirect(w, r, t.PageLocation(), http.StatusFound)
					}
				}, w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}
	}
	s.NotFound(w, map[string]interface{}{
		"Error": "torrent not found",
	}, j)
}

func (s *Server) checkAuth(r *http.Request) (ok, requireCaptcha bool, err error) {
	var user, passwd string
	user, passwd, ok = r.BasicAuth()
	if ok {
		ok, err = s.DB.CheckLogin(user, passwd)
	} else {
		sol := r.FormValue("captcha-solution")
		id := r.FormValue("captcha-id")
		ok = captcha.VerifyString(id, sol)
		requireCaptcha = !ok
	}
	return
}

func (s *Server) requireAuth(handler http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	j := s.shouldJSON(r)
	a := s.shouldAUTH(r)
	ok, wantCaptcha, err := s.checkAuth(r)
	if ok {
		handler(w, r)
	} else if a {
		w.Header().Set("WWW-Authenticate", "Basic")
		w.WriteHeader(http.StatusUnauthorized)
	} else if wantCaptcha {
		s.Error(w, "invalid captcha", j)
	} else if err == nil {
		w.Header().Set("WWW-Authenticate", "Basic")
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		s.Error(w, err.Error(), j)
	}
}

func New(cfg *config.IndexConfig) (s *Server) {

	funcs := template.FuncMap{
		// ISO 8601 human readable version
		"FormatDate": func(t time.Time) string {
			t = t.UTC()
			Y, M, D := t.Date()
			h, m, s := t.Clock()
			return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", Y, M, D, h, m, s)
		},
		// machine readable, manually specify global timezone
		"FormatDateGlobal": func(t time.Time) string {
			t = t.UTC()
			Y, M, D := t.Date()
			h, m, s := t.Clock()
			return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ", Y, M, D, h, m, s)
		},
		// RFC 2822 Date: header style
		"FormatDateRFC2822": func(t time.Time) string {
			t = t.UTC()
			Y, M, D := t.Date()
			W := t.Weekday()
			h, m, s := t.Clock()
			return fmt.Sprintf("%s, %02d %s %04d %02d:%02d:%02d", W, D, M, Y, h, m, s)
		},
	}
	s = &Server{
		cfg:  cfg,
		mux:  http.NewServeMux(),
		tmpl: template.Must(template.New("").Funcs(funcs).ParseGlob(filepath.Join(cfg.TemplateDir, "**"))),
	}

	// set up routes
	s.mux.Handle("/static/", http.FileServer(http.Dir(cfg.StaticDir)))
	s.mux.Handle("/captcha/", captcha.Server(cfg.CaptchaWidth, cfg.CaptchaHeight))
	s.mux.HandleFunc("/c/", s.handleCategoryPage)
	s.mux.HandleFunc("/dl/", s.serveTorrent)
	s.mux.HandleFunc("/t/", s.serveTorrentInfo)
	s.mux.HandleFunc("/s/", s.handleSearch)
	s.mux.HandleFunc("/", s.serveFrontPage)
	return
}
