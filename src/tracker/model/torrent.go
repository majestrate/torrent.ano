package model

import (
	"encoding/hex"
	"fmt"
	"time"
)

type File struct {
	Name string
	Size uint64
}

type Torrent struct {
	Name         string
	Tags         []Tag
	Category     Category
	PieceSize    uint32
	Size         uint64
	Uploaded     int64
	IH           [20]byte
	AnnounceURLS []string
	GetFiles     func() []File
}

func (t *Torrent) UploadedAt() time.Time {
	return time.Unix(t.Uploaded, 0)
}

func (t *Torrent) InfoHash() string {
	return hex.EncodeToString(t.IH[:])
}

func (t *Torrent) DownloadLink() string {
	return fmt.Sprintf("/dl/%s.torrent", t.InfoHash())
}

func (t *Torrent) Magnet() string {
	trs := ""
	if t.AnnounceURLS != nil {
		for _, tr := range t.AnnounceURLS {
			trs += fmt.Sprintf("&tr=%s", tr)
		}
	}
	return fmt.Sprintf("magnet:?xt=urn:btih:%s%s", t.InfoHash(), trs)
}
