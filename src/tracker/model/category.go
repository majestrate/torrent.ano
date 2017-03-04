package model

type Category struct {
	Name        string
	ID          int
	GetTorrents func() []Torrent
}
