package model

import (
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"
	"tracker/util"
)

type File struct {
	Name string
	Size uint64
}

func (f *File) SizeString() string {
	return util.SizeString(f.Size)
}

type Torrent struct {
	Domain       string
	Name         string
	Tags         []Tag
	Category     Category
	PieceSize    uint32
	Size         uint64
	Uploaded     int64
	IH           [20]byte
	AnnounceURLS []string
}

func (t *Torrent) MarshalJSON() (data []byte, err error) {
	id := t.Category.ID
	m := map[string]interface{}{
		"Name":        t.Name,
		"PieceSize":   t.PieceSize,
		"Size":        t.Size,
		"Uploaded":    t.Uploaded,
		"InfoHash":    t.InfoHash(),
		"DownloadURL": NewLink(t.Domain, t.DownloadLink(), "").URL,
		"Magnet":      t.Magnet(),
		"InfoURL":     NewLink(t.Domain, fmt.Sprintf("/t/%s/?t=json", t.InfoHash()), "").URL,
	}
	if id != 0 {
		m["Category"] = id
	}
	data, err = json.Marshal(m)
	return
}

func (t *Torrent) SizeString() string {
	return util.SizeString(t.Size)
}

func (t Torrent) UploadedAt() time.Time {
	return time.Unix(t.Uploaded, 0)
}

func (t *Torrent) InfoHash() string {
	return hex.EncodeToString(t.IH[:])
}

func (t *Torrent) PageLocation() string {
	return fmt.Sprintf("/t/%s", t.InfoHash())
}

func (t *Torrent) DownloadLink() string {
	return fmt.Sprintf("/dl/%s.torrent", t.InfoHash())
}

func (t *Torrent) Magnet() string {
	return fmt.Sprintf("magnet:?xt=urn:btih:%s", t.InfoHash())
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
		Title:   t.Name,
		Link:    NewLink(t.Domain, t.PageLocation(), ""),
		ID:      t.InfoHash(),
		Updated: t.UploadedAt(),
		Summary: ("Torrent: <a href=\"" + (t.Domain + t.DownloadLink()) + "\">" + (t.Domain + t.DownloadLink()) + "</a> Infohash: " + t.InfoHash() + " [Size: " + t.SizeString() + "]"),
		//		AuthorName: "Anonymous Uploader",
	}
}

func (t Torrent) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	err = e.EncodeElement(t.toFeed(), start)
	return
}
