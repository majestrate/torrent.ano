package db

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/lib/pq"
	"tracker/config"
	"tracker/log"
	"tracker/model"
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
	}

	for _, name := range tableOrder {
		err = st.createTable(name, tables[name])
		if err != nil {
			log.Errorf("failed to make table %s: %s", name, err)
			break
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
	err = st.conn.QueryRow(fmt.Sprintf("SELECT name, uploaded_at, pieces_size, total_size FROM %s WHERE infohash = $1 LIMIT 1", tableMetaInfo), i).Scan(&t.Name, &t.Uploaded, &t.PieceSize, &t.Size)
	if err == sql.ErrNoRows {
		err = nil
		t = nil
	}
	return
}

func (st *Postgres) GetAllCategories() (cats []model.Category, err error) {
	var rows *sql.Rows
	rows, err = st.conn.Query(fmt.Sprintf("SELECT name, id FROM %s", tableCategories))
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

func (st *Postgres) FindTorrentsInCategory(cat *model.Category) (torrents []model.Torrent, err error) {
	var rows *sql.Rows
	rows, err = st.conn.Query(fmt.Sprintf("SELECT name, uploaded_at, pieces_size, total_size, infohash FROM %s WHERE category_id = $1", tableMetaInfo), cat.ID)
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

func (st *Postgres) FindTorrentsWithTags(tags []model.Tag) (torrents []model.Torrent, err error) {
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
	return
}

func (st *Postgres) GetTagByName(name string) (tag *model.Tag, err error) {
	return
}

func (st *Postgres) StoreTorrent(t *model.Torrent) (err error) {
	ih := t.InfoHash()
	// insert torrent meta info
	_, err = st.conn.Exec(fmt.Sprintf("INSERT INTO %s(infohash, pieces_size, total_size, name, uploaded_at, category_id) VALUES ($1, $2, $3, $4, $5, $6)", tableMetaInfo), ih, t.PieceSize, t.Size, t.Name, t.Uploaded, t.Category.ID)
	if err != nil {
		return
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
