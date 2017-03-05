package db

import (
	"tracker/model"
)

// DB defines actions required by a database driver
type DB interface {
	Init() error
	StoreTorrent(*model.Torrent) error
	FindTorrentByInfohash(ih [20]byte) (*model.Torrent, error)
	FindTorrentsWithTags(tags []model.Tag) ([]model.Torrent, error)
	GetTagByName(name string) (*model.Tag, error)
	GetTagByID(id uint64) (*model.Tag, error)
	GetCategoryByID(id int) (*model.Category, error)
	FindTorrentsInCategory(*model.Category) ([]model.Torrent, error)
	GetAllCategories() ([]model.Category, error)
	GetFrontPageTorrents() ([]model.Torrent, error)
}
