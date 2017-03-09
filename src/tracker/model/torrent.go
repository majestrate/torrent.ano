package model

import (
	"encoding/hex"
	"encoding/xml"
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

type torrentFeed struct {
	Title      string    `xml:"title"`
	Link       Link      `xml:"link"`
	ID         string    `xml:"id"`
	Updated    time.Time `xml:"updated"`
	Summary    string    `xml:"summary"`
	AuthorName string    `xml:"author>name"`
}

func (t *Torrent) toFeed() *torrentFeed {
	return &torrentFeed{
		Title:      t.Name,
		Link:       Link{t.DownloadLink()},
		ID:         t.InfoHash(),
		Updated:    t.UploadedAt(),
		Summary:    t.Name,
		AuthorName: "anonymous uploader",
	}
}

func (t *Torrent) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	err = e.EncodeElement(t.toFeed(), start)
	return
}
