package db

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/majestrate/torrent.ano/lib/tracker/config"
	"github.com/majestrate/torrent.ano/lib/tracker/log"
	"github.com/majestrate/torrent.ano/lib/tracker/metainfo"
	"github.com/majestrate/torrent.ano/lib/tracker/model"
	"github.com/majestrate/torrent.ano/lib/tracker/util"
	"html"
	"strings"
	"time"
)

// PQTorrentStorage is a postgresql torrent metadata storage implementation
type Postgres struct {
	conn  *sql.DB
	stmts map[string]*sql.Stmt
}

const tableCategories = "Categories"
const tableTags = "UserTags"
const tableMetaInfo = "MetaInfos"
const tableTagMetaInt = "MetaInfoTags"
const tableAnnouncers = "Announcers"
const tableAnnouncerMetaInfoInt = "MetaInfoAnnouncers"
const tableFiles = "MetaInfoFiles"
const tableSwarmEvents = "SwarmEvents"
const tableAuthedUsers = "AuthedUsers"
const tableComments = "TorrentComments"

func (st *Postgres) ensureTables() (err error) {
	tables := map[string]string{
		tableCategories:           "( name VARCHAR(255) NOT NULL, id SERIAL PRIMARY KEY )",
		tableTags:                 "( name VARCHAR(255) NOT NULL, id SERIAL PRIMARY KEY )",
		tableMetaInfo:             fmt.Sprintf("( infohash VARCHAR(40) PRIMARY KEY, pieces_size BIGINT NOT NULL, total_size BIGINT NOT NULL, name VARCHAR(512) NOT NULL, uploaded_at BIGINT NOT NULL, category_id INTEGER REFERENCES %s (id) ON DELETE CASCADE )", tableCategories),
		tableFiles:                fmt.Sprintf("( id SERIAL PRIMARY KEY, filename VARCHAR(512) NOT NULL, filesize BIGINT, meta_infohash VARCHAR(40) REFERENCES %s(infohash) ON DELETE CASCADE, UNIQUE (filename, meta_infohash) ) ", tableMetaInfo),
		tableAnnouncers:           "( domain VARCHAR(255) NOT NULL, protocol VARCHAR(18) NOT NULL, port INTEGER NOT NULL, path VARCHAR(255) NOT NULL, id SERIAL PRIMARY KEY )",
		tableTagMetaInt:           fmt.Sprintf("( tag_id BIGINT REFERENCES %s(id) ON DELETE CASCADE, tag_infohash VARCHAR(40) REFERENCES %s(infohash) ON DELETE CASCADE, UNIQUE(tag_id, tag_infohash) )", tableTags, tableMetaInfo),
		tableAnnouncerMetaInfoInt: fmt.Sprintf("( announce_id INTEGER REFERENCES %s(id) ON DELETE RESTRICT, meta_infohash  VARCHAR(40) REFERENCES %s(infohash) ON DELETE RESTRICT, UNIQUE (announce_id, meta_infohash) )", tableAnnouncers, tableMetaInfo),
		tableSwarmEvents:          fmt.Sprintf("( swarm_infohash VARCHAR(40) REFERENCES %s(infohash) ON DELETE CASCADE, seeders INTEGER NOT NULL, leechers INTEGER NOT NULL, event_at BIGINT NOT NULL )", tableMetaInfo),
		tableAuthedUsers:          "( username VARCHAR(255) NOT NULL, credential VARCHAR(255), id SERIAL PRIMARY KEY)",
		tableComments:             fmt.Sprintf("( comment_infohash VARCHAR(40) REFERENCES %s(infohash) ON DELETE CASCADE, message TEXT NOT NULL, posted BIGINT NOT NULL, id SERIAL PRIMARY KEY )", tableMetaInfo),
	}

	tableOrder := []string{
		tableCategories,
		tableTags,
		tableMetaInfo,
		tableSwarmEvents,
		tableAnnouncers,
		tableFiles,
		tableAnnouncerMetaInfoInt,
		tableTagMetaInt,
		tableAuthedUsers,
		tableComments,
	}

	for _, name := range tableOrder {
		err = st.createTable(name, tables[name])
		if err != nil {
			log.Errorf("failed to make table %s: %s", name, err)
			break
		}
	}
	postQueries := []string{
		`CREATE OR REPLACE FUNCTION tagrank(tid integer) RETURNS BIGINT AS 'SELECT COUNT(DISTINCT metainfotags.tag_infohash) AS rank FROM metainfotags WHERE tag_id = tid;' LANGUAGE SQL IMMUTABLE RETURNS NULL ON NULL INPUT`,
	}
	for _, f := range postQueries {
		_, err = st.conn.Exec(f)
		if err != nil {
			log.Errorf("failed to execute post query: %s", err.Error())
		}
	}
	return
}

// create a table given name and table defintion
func (st *Postgres) createTable(name, def string) (err error) {
	_, err = st.conn.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s %s", name, def))
	return
}

func (st *Postgres) prepareStatements() (err error) {

	return
}

func (st *Postgres) Init() (err error) {
	err = st.ensureTables()
	if err == nil {
		err = st.prepareStatements()
	}
	return
}

func (st *Postgres) GetFrontPageTorrents() (torrents []model.Torrent, err error) {
	var rows *sql.Rows
	rows, err = st.conn.Query(fmt.Sprintf("SELECT infohash, name, uploaded_at, total_size, category_id FROM %s ORDER BY uploaded_at DESC LIMIT 10", tableMetaInfo))
	if err == nil {
		for rows.Next() {
			var t model.Torrent
			var ih string
			rows.Scan(&ih, &t.Name, &t.Uploaded, &t.Size, &t.Category.ID)
			infohash, _ := hex.DecodeString(ih)
			copy(t.IH[:], infohash)
			cat, _ := st.GetCategoryByID(t.Category.ID)
			t.Category.Name = cat.Name
			torrents = append(torrents, t)
		}
		rows.Close()
	} else if err == sql.ErrNoRows {
		err = nil
	}

	return
}

func (st *Postgres) FindTorrentByInfohash(ih [20]byte) (t *model.Torrent, err error) {
	i := hex.EncodeToString(ih[:])
	t = new(model.Torrent)
	copy(t.IH[:], ih[:])
	err = st.conn.QueryRow(fmt.Sprintf("SELECT name, uploaded_at, pieces_size, total_size, category_id FROM %s WHERE infohash = $1 LIMIT 1", tableMetaInfo), i).Scan(&t.Name, &t.Uploaded, &t.PieceSize, &t.Size, &t.Category.ID)
	if err == sql.ErrNoRows {
		err = nil
		t = nil
	}
	return
}

func (st *Postgres) GetAllCategories() (cats []model.Category, err error) {
	var rows *sql.Rows
	rows, err = st.conn.Query(fmt.Sprintf("SELECT name, id FROM %s ORDER BY name", tableCategories))
	if err == nil {
		for rows.Next() {
			var cat model.Category
			rows.Scan(&cat.Name, &cat.ID)
			cats = append(cats, cat)
		}
		rows.Close()
	}
	return

}

func (st *Postgres) FindTorrentsInCategory(cat *model.Category, perpage, offset int) (torrents []model.Torrent, err error) {
	var rows *sql.Rows
	rows, err = st.conn.Query(fmt.Sprintf("SELECT name, uploaded_at, pieces_size, total_size, infohash FROM %s WHERE category_id = $1 ORDER BY uploaded_at DESC LIMIT $2 OFFSET $3", tableMetaInfo), cat.ID, perpage, offset)
	if err == nil {
		for rows.Next() {
			var t model.Torrent
			var ih string
			rows.Scan(&t.Name, &t.Uploaded, &t.PieceSize, &t.Size, &ih)
			b, _ := hex.DecodeString(ih)
			copy(t.IH[:], b[:])
			torrents = append(torrents, t)
		}
		rows.Close()
	}
	return
}

func (st *Postgres) GetTorrentFiles(ih [20]byte) (files []model.File, err error) {
	ihstr := hex.EncodeToString(ih[:])
	var rows *sql.Rows
	rows, err = st.conn.Query(fmt.Sprintf("SELECT filename, filesize FROM %s WHERE meta_infohash = $1 ORDER BY filename", tableFiles), ihstr)
	if err == nil {
		for rows.Next() {
			var f model.File
			rows.Scan(&f.Name, &f.Size)
			files = append(files, f)
		}
		rows.Close()
	}
	return
}

func (st *Postgres) ListPopularTags(limit int) (tags []model.Tag, err error) {
	var rows *sql.Rows
	rows, err = st.conn.Query(fmt.Sprintf("SELECT tagrank(id) as rank, id, name FROM %s ORDER BY rank DESC LIMIT $1", tableTags), limit)
	if err == nil {
		for rows.Next() {
			var tag model.Tag
			rows.Scan(&tag.Rank, &tag.ID, &tag.Name)
			if len(tag.Name) > 0 {
				tags = append(tags, tag)
			}
		}
	} else if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (st *Postgres) FindTorrentsWithTag(tag model.Tag) (torrents []model.Torrent, err error) {
	var rows *sql.Rows
	rows, err = st.conn.Query(fmt.Sprintf("SELECT i.infohash, i.name, i.uploaded_at, i.pieces_size, i.total_size FROM %s i INNER JOIN ( SELECT tag_infohash FROM %s WHERE tag_id = $1 ) t ON t.tag_infohash = i.infohash ORDER BY i.uploaded_at DESC", tableMetaInfo, tableTagMetaInt), tag.ID)
	if err == nil {
		for rows.Next() {
			var ih string
			var torrent model.Torrent
			rows.Scan(&ih, &torrent.Name, &torrent.Uploaded, &torrent.PieceSize, &torrent.Size)
			d, _ := hex.DecodeString(ih)
			copy(torrent.IH[:], d)
			torrents = append(torrents, torrent)
		}
	} else if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func (st *Postgres) GetCategoryByID(id int) (cat *model.Category, err error) {
	cat = &model.Category{
		ID: id,
	}
	err = st.conn.QueryRow(fmt.Sprintf("SELECT name FROM %s WHERE id = $1 LIMIT 1", tableCategories), id).Scan(&cat.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		cat = nil
	}
	return
}

func (st *Postgres) GetTagByID(id uint64) (tag *model.Tag, err error) {
	tag = &model.Tag{
		ID: id,
	}
	err = st.conn.QueryRow(fmt.Sprintf("SELECT name, tagrank(id) FROM %s WHERE id = $1 LIMIT 1", tableTags), id).Scan(&tag.Name, &tag.Rank)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		tag = nil
	}
	return
}

func (st *Postgres) GetTagByName(name string) (tag *model.Tag, err error) {
	tag = &model.Tag{
		Name: name,
	}
	err = st.conn.QueryRow(fmt.Sprintf("SELECT id, tagrank(id) FROM %s WHERE name = $1 LIMIT 1", tableTags), name).Scan(&tag.ID, &tag.Rank)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		tag = nil
	}
	return
}

func (st *Postgres) StoreTorrent(t *model.Torrent, i *metainfo.TorrentFile) (err error) {
	ih := t.InfoHash()
	// insert torrent meta info
	_, err = st.conn.Exec(fmt.Sprintf("INSERT INTO %s (infohash, pieces_size, total_size, name, uploaded_at, category_id) VALUES ($1, $2, $3, $4, $5, $6)", tableMetaInfo), ih, t.PieceSize, t.Size, t.Name, t.Uploaded, t.Category.ID)
	if err != nil {
		return
	}

	// insert files
	files := i.Info.GetFiles()
	for _, f := range files {
		_, err = st.conn.Exec(fmt.Sprintf("INSERT INTO %s (meta_infohash, filesize, filename) VALUES ( $1, $2, $3 )", tableFiles), ih, f.Length, f.Path.FilePath())
		if err != nil {
			return
		}
	}

	// insert tags
	for _, tag := range t.Tags {
		if tag.ID == 0 {
			// no tag id set
			var count int
			// check for existing tag
			err = st.conn.QueryRow(fmt.Sprintf("SELECT COUNT(id) FROM %s WHERE name = $1", tableTags), tag.Name).Scan(&count)
			if err != nil {
				return
			}
			if count <= 0 {
				// new tag
				_, err = st.conn.Exec(fmt.Sprintf("INSERT INTO %s (name) VALUES ($1)", tableTags), tag.Name)
			}
			// get tag id
			err = st.conn.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE name = $1 LIMIT 1", tableTags), tag.Name).Scan(&tag.ID)
			if err != nil {
				return
			}
		}
		// insert tag -> metainfo entry
		_, err = st.conn.Exec(fmt.Sprintf("INSERT INTO %s (tag_id, tag_infohash) VALUES ($1, $2)", tableTagMetaInt), tag.ID, ih)
		if err != nil {
			return
		}
	}
	return
}

func (db *Postgres) CheckLogin(user, passwd string) (authed bool, err error) {
	var u model.User
	err = db.conn.QueryRow(fmt.Sprintf("SELECT username, credential FROM %s WHERE username = $1 LIMIT 1", tableAuthedUsers), user).Scan(&u.Username, &u.Login)
	if err == sql.ErrNoRows {
		err = nil
	} else if err == nil {
		authed = u.Login.Check(passwd)
	}
	return
}

func (db *Postgres) AddUserLogin(user, passwd string) (err error) {
	var count int
	err = db.conn.QueryRow(fmt.Sprintf("SELECT COUNT(username) FROM %s WHERE username = $1", tableAuthedUsers), user).Scan(&count)
	if err == nil {
		if count > 0 {
			err = ErrUserExists
		} else {
			u := model.NewUser(user, passwd)
			_, err = db.conn.Exec(fmt.Sprintf("INSERT INTO %s (username, credential) VALUES ($1, $2)", tableAuthedUsers), u.Username, u.Login.String())
		}
	}
	return
}

func (db *Postgres) DelUserLogin(user string) (err error) {
	_, err = db.conn.Exec(fmt.Sprintf("DELETE FROM %s WHERE username = $1", tableAuthedUsers), user)
	return
}

func (db *Postgres) InsertComment(text string, ih [20]byte) (err error) {
	text = html.EscapeString(text)
	infohash := hex.EncodeToString(ih[:])
	now := time.Now().Unix()
	_, err = db.conn.Exec(fmt.Sprintf("INSERT INTO %s(comment_infohash, message, posted) VALUES($1, $2, $3)", tableComments), infohash, text, now)
	return
}

func (db *Postgres) GetCommentsForTorrent(t *model.Torrent) (comments []model.Comment, err error) {
	var rows *sql.Rows
	rows, err = db.conn.Query(fmt.Sprintf("SELECT message, posted, id FROM %s WHERE comment_infohash = $1 ORDER BY posted ASC", tableComments), t.InfoHash())
	if err == sql.ErrNoRows {
		err = nil
	} else if err == nil {
		for rows.Next() {
			var c model.Comment
			var posted int64
			rows.Scan(&c.Text, &posted, &c.ID)
			c.Posted = time.Unix(posted, 0)
			comments = append(comments, c)
		}
		rows.Close()
	}
	return
}

func (db *Postgres) GetTorrentTags(t *model.Torrent) (tags []model.Tag, err error) {
	var rows *sql.Rows
	rows, err = db.conn.Query(fmt.Sprintf("SELECT id, name, tagrank(id) as rank FROM %s WHERE id IN ( SELECT tag_id from %s WHERE tag_infohash = $1) ORDER BY rank DESC", tableTags, tableTagMetaInt), t.InfoHash())
	if err == sql.ErrNoRows {
		err = nil
	} else if err == nil {
		for rows.Next() {
			var tag model.Tag
			rows.Scan(&tag.ID, &tag.Name, &tag.Rank)
			tags = append(tags, tag)
		}
		rows.Close()
	}
	return
}

func (db *Postgres) EnsureTags(tags []string) (ensured []model.Tag, err error) {
	ensured = make([]model.Tag, len(tags))
	for idx, name := range tags {
		name = strings.TrimFunc(name, util.IsSpace)
		if len(name) == 0 {
			continue
		}
		var tag *model.Tag
		tag, err = db.GetTagByName(name)
		if tag == nil && err == nil {
			ensured[idx].Name = name
			// new tag
			_, err = db.conn.Exec(fmt.Sprintf("INSERT INTO %s (name) VALUES ($1)", tableTags), ensured[idx].Name)
			// get tag id
			err = db.conn.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE name = $1 LIMIT 1", tableTags), ensured[idx].Name).Scan(&ensured[idx].ID)
			if err != nil {
				ensured = nil
				return
			}
		} else if err != nil {
			ensured = nil
			return
		} else {
			ensured[idx].Name = tag.Name
			ensured[idx].ID = tag.ID
		}
	}
	return
}

func (db *Postgres) AddTorrentTags(tags []model.Tag, t *model.Torrent) (err error) {
	if len(tags) == 0 {
		return
	}
	ih := t.InfoHash()
	var tagValues string
	if len(tags) > 1 {
		for idx, tag := range tags {
			if idx == 0 {
				tagValues += fmt.Sprintf("('%s', %d ) ", ih, tag.ID)
			} else {
				tagValues += fmt.Sprintf(", ('%s', %d ) ", ih, tag.ID)
			}
		}
	} else {
		tagValues = fmt.Sprintf("( '%s', %d )", ih, tags[0].ID)
	}
	_, err = db.conn.Exec(fmt.Sprintf("INSERT INTO %s(tag_infohash, tag_id) VALUES %s", tableTagMetaInt, tagValues))
	return
}

func (db *Postgres) DelTorrentTags(tags []model.Tag, t *model.Torrent) (err error) {
	var tagValues string
	ih := t.InfoHash()
	for idx, tag := range tags {
		if idx == 0 {
			tagValues += fmt.Sprintf("( tag_infohash = '%s' AND tag_id = %d ) ", ih, tag.ID)
		} else {
			tagValues += fmt.Sprintf("OR ( tag_infohash = '%s' AND tag_id = %d ) ", ih, tag.ID)
		}
	}
	_, err = db.conn.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s", tableTagMetaInt, tagValues))
	return
}

func (db *Postgres) DelTorrent(ih string) (err error) {
	_, err = db.conn.Exec(fmt.Sprintf("DELETE FROM %s WHERE infohash = $1", tableMetaInfo), ih)
	return
}

func (db *Postgres) AddCategory(name string) (err error) {
	_, err = db.conn.Exec(fmt.Sprintf("INSERT INTO %s(name) VALUES ($1)", tableCategories), name)
	return
}

func (db *Postgres) DelCategory(name string) (err error) {
	_, err = db.conn.Exec(fmt.Sprintf("DELETE FROM %s WHERE name = $1", tableCategories), name)
	return
}

func (db *Postgres) HasTorrent(ih [20]byte) (has bool, err error) {
	ihhex := hex.EncodeToString(ih[:])
	var rows *sql.Rows
	rows, err = db.conn.Query(fmt.Sprintf("SELECT infohash from %s WHERE infohash = $1", tableMetaInfo), ihhex)
	if err == nil {
		has = true
		rows.Close()
	} else if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func NewPostgres(cfg *config.DBConfig) (db *Postgres, err error) {
	var conn *sql.DB
	conn, err = sql.Open("postgres", cfg.URL)
	if err == nil {
		db = &Postgres{
			stmts: make(map[string]*sql.Stmt),
			conn:  conn,
		}
	}
	return
}
