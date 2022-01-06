package storage

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/majestrate/torrent.ano/lib/tracker/config"
)

// PQTorrentStorage is a postgresql torrent metadata storage implementation
type PQTorrentStorage struct {
	conn *sql.DB
}

func NewPQ(cfg *config.DBConfig) (db *PQTorrentStorage, err error) {
	db = new(PQTorrentStorage)
	db.conn, err = sql.Open(cfg.Type, cfg.URL)
	return
}
