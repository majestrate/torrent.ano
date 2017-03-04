package index

import (
	"fmt"
	"html/template"
	"net/http"
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
)

const frontPage = "frontpage.html.tmpl"
const categoryPage = "torrent-category.html.tmpl"

type Server struct {
	cfg  *config.IndexConfig
	mux  *http.ServeMux
	tmpl *template.Template
	DB   db.DB
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
func (s *Server) Error(w http.ResponseWriter, msg string) {
	s.tmpl.ExecuteTemplate(w, "error.html.tmpl", map[string]interface{}{
		"Error": msg,
	})
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

func (s *Server) makeParams() map[string]interface{} {
	return map[string]interface{}{}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) addTorrent(w http.ResponseWriter, r *http.Request, cat model.Category, p map[string]interface{}) {
	store := s.DB
	if store == nil {
		s.Error(w, "internal error: no storage backend")
		return
	}
	defer r.Body.Close()
	sol := r.FormValue("captcha-solution")
	id := r.FormValue("captcha-id")
	tags := r.FormValue("torrent-tags")
	name := r.FormValue("torrent-name")
	if captcha.VerifyString(id, sol) {
		t := new(metainfo.TorrentFile)
		f, h, err := r.FormFile("torrent-file")
		if name == "" {
			name = h.Filename
		}
		if name == "" {
			s.Error(w, "torrent name not specified")
			return
		}
		if err != nil {
			s.Error(w, err.Error())
			return
		}
		err = t.Decode(f)
		if err != nil {
			s.Error(w, err.Error())
			return
		}

		if t.Info.Private > 0 {
			s.Error(w, "private torrents not allowed")
			return
		}

		t = s.filterTorrent(t)

		torrent, err := store.FindTorrentByInfohash(t.Infohash())
		if err != nil {
			s.Error(w, err.Error())
			return
		}
		if torrent != nil {
			s.Error(w, "duplicate torrent")
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
		for _, tag := range strings.Split(tags, ",") {
			tname := strings.Replace(strings.Trim(tag, " "), " ", "-", -1)
			torrent.Tags = append(torrent.Tags, model.Tag{
				Name: tname,
			})
		}
		err = store.StoreTorrent(torrent)
		if err != nil {
			s.Error(w, "could not store torrent: "+err.Error())
			return
		}
		// store torrent seed file
		fpath := filepath.Join(s.cfg.TorrentsDir, fmt.Sprintf("%s.torrent", torrent.InfoHash()))
		var file *os.File
		file, err = os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			s.Error(w, "could not open file: "+err.Error())
			return
		}
		err = t.Encode(file)
		file.Close()
		if err != nil {
			s.Error(w, err.Error())
			return
		}
		p["Torrent"] = name
		// success
		w.Header().Set("Location", fmt.Sprintf("/t/%s/", torrent.InfoHash()))
		w.WriteHeader(http.StatusFound)
		s.tmpl.ExecuteTemplate(w, "success.html.tmpl", p)
	} else {
		s.Error(w, "bad captcha")
	}
}

func (s *Server) NotFound(w http.ResponseWriter, p map[string]interface{}) {
	w.WriteHeader(http.StatusNotFound)
	s.tmpl.ExecuteTemplate(w, "not-found.html.tmpl", p)
}

func (s *Server) handleCategoryPage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	p := map[string]interface{}{
		"Message": "No Such Category",
	}
	catid, err := strconv.Atoi(strings.Trim(path[3:], "/"))
	if err != nil {
		s.NotFound(w, p)
		return
	}
	cat, err := s.DB.GetCategoryByID(catid)
	if err != nil {
		s.Error(w, err.Error())
		return
	}
	if cat == nil {
		s.NotFound(w, p)
		return
	}

	if r.Method == "GET" {
		torrents, err := s.DB.FindTorrentsInCategory(cat)
		if err != nil {
			s.Error(w, err.Error())
			return
		}
		err = s.tmpl.ExecuteTemplate(w, "category.html.tmpl", map[string]interface{}{
			"Torrents": torrents,
			"Category": cat,
			"Captcha":  captcha.New(),
		})
		if err != nil {
			log.Errorf("failed to render category page: %s", err)
		}
	} else if r.Method == "POST" {
		s.addTorrent(w, r, *cat, s.makeParams())
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

	cats, err := s.DB.GetAllCategories()
	if err != nil {
		s.Error(w, "failed to fetch categories: "+err.Error())
		return
	}
	err = s.tmpl.ExecuteTemplate(w, "frontpage.html.tmpl", map[string]interface{}{
		"Categories": cats,
	})
	if err != nil {
		log.Errorf("failed to render front page: %s", err)
	}
}

func New(cfg *config.IndexConfig) (s *Server) {
	s = &Server{
		cfg:  cfg,
		mux:  http.NewServeMux(),
		tmpl: template.Must(template.ParseGlob(filepath.Join(cfg.TemplateDir, "**"))),
	}

	// set up routes
	s.mux.Handle("/static/", http.FileServer(http.Dir(cfg.StaticDir)))
	s.mux.Handle("/captcha/", captcha.Server(cfg.CaptchaWidth, cfg.CaptchaHeight))
	s.mux.HandleFunc("/c/", s.handleCategoryPage)
	s.mux.HandleFunc("/dl/", s.serveTorrent)
	s.mux.HandleFunc("/", s.serveFrontPage)
	s.mux.HandleFunc("/search/", s.handleSearch)
	return
}
