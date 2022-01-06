package storage

import (
	"github.com/majestrate/torrent.ano/lib/tracker/model"
)

// TorrentStorage stores torrent metadata to a backend
type TorrentStorage interface {
	StoreTorrent(t *model.Torrent) error
}
