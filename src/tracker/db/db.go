package db

import (
	"errors"
	"tracker/metainfo"
	"tracker/model"
)

var ErrUserExists = errors.New("user already exists")

// DB defines actions required by a database driver
type DB interface {
	Init() error
	StoreTorrent(*model.Torrent, *metainfo.TorrentFile) error
	FindTorrentByInfohash(ih [20]byte) (*model.Torrent, error)
	FindTorrentsWithTag(tag model.Tag) ([]model.Torrent, error)
	ListPopularTags(limit int) ([]model.Tag, error)
	GetTagByName(name string) (*model.Tag, error)
	GetTagByID(id uint64) (*model.Tag, error)
	GetCategoryByID(id int) (*model.Category, error)
	FindTorrentsInCategory(*model.Category) ([]model.Torrent, error)
	GetAllCategories() ([]model.Category, error)
	GetFrontPageTorrents() ([]model.Torrent, error)
	GetTorrentFiles(ih [20]byte) ([]model.File, error)
	CheckLogin(user, password string) (bool, error)
	AddUserLogin(username, password string) error
	DelUserLogin(username string) error
}
