package storage

import (
	"tracker/model"
)

// TorrentStorage stores torrent metadata to a backend
type TorrentStorage interface {
	StoreTorrent(t *model.Torrent) error
}
