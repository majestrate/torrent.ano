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
	HasTorrent(ih [20]byte) (bool, error)
	FindTorrentByInfohash(ih [20]byte) (*model.Torrent, error)
	FindTorrentsWithTag(tag model.Tag) ([]model.Torrent, error)
	ListPopularTags(limit int) ([]model.Tag, error)
	GetTagByName(name string) (*model.Tag, error)
	GetTagByID(id uint64) (*model.Tag, error)
	GetCategoryByID(id int) (*model.Category, error)
	FindTorrentsInCategory(*model.Category, int, int) ([]model.Torrent, error)
	FindTorrentsByFile(name string, perpage, offset int) ([]model.Torrent, error)
	GetAllCategories() ([]model.Category, error)
	GetFrontPageTorrents() ([]model.Torrent, error)
	GetTorrentFiles(ih [20]byte) ([]model.File, error)
	CheckLogin(user, password string) (bool, error)
	AddUserLogin(username, password string) error
	DelUserLogin(username string) error
	InsertComment(text string, ih [20]byte) error
	GetCommentsForTorrent(*model.Torrent) ([]model.Comment, error)
	GetTorrentTags(*model.Torrent) ([]model.Tag, error)
	EnsureTags(tags []string) ([]model.Tag, error)
	AddTorrentTags(tags []model.Tag, t *model.Torrent) error
	DelTorrentTags(tags []model.Tag, t *model.Torrent) error
	AddCategory(name string) error
	DelCategory(name string) error
	DelTorrent(ih string) error
}
