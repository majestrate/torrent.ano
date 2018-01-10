package model

import (
	"encoding/xml"
	"fmt"
	"time"
)

type Comment struct {
	ID      uint64
	Text    string
	Posted  time.Time
	Domain  string
	Torrent *Torrent
}

type commentFeed struct {
	Title      string    `xml:"title"`
	Link       Link      `xml:"link"`
	ID         string    `xml:"id"`
	Updated    time.Time `xml:"updated"`
	Summary    string    `xml:"summary"`
	AuthorName string    `xml:"author>name"`
}

func (c Comment) UploadedAt() time.Time {
	return c.Posted
}

func (c *Comment) toFeed() *commentFeed {

	return &commentFeed{
		Title:      "comment",
		Link:       NewLink(c.Domain, fmt.Sprintf("%s#comment_%d", c.Torrent.PageLocation(), c.ID)),
		ID:         fmt.Sprintf("%s-%d", c.Torrent.InfoHash(), c.ID),
		Updated:    c.UploadedAt(),
		Summary:    "new comment",
		AuthorName: "anonymous",
	}
}

func (c Comment) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	err = e.EncodeElement(c.toFeed(), start)
	return
}
