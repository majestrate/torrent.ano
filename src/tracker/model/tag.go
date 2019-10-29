package model

import (
	"encoding/xml"
	"fmt"
	"time"
)

type Tag struct {
	Name   string
	ID     uint64
	Rank   uint64
	Domain string
}

func (t *Tag) SearchLink() string {
	return fmt.Sprintf("/s/?id=%d", t.ID)
}

type tagFeed struct {
	Title      string    `xml:"title"`
	Link       Link      `xml:"link"`
	ID         string    `xml:"id"`
	Updated    time.Time `xml:"updated"`
	Summary    string    `xml:"summary"`
	AuthorName string    `xml:"author>name"`
}

func (t Tag) UploadedAt() time.Time {
	return time.Now()
}

func (t *Tag) toFeed() *tagFeed {

	return &tagFeed{
		Title:      t.Name,
		Link:       NewQueryLink(t.Domain, "/s/", fmt.Sprintf("id=%d", t.ID)),
		ID:         fmt.Sprintf("tag-%d", t.ID),
		Updated:    t.UploadedAt(),
		Summary:    t.Name,
		AuthorName: "anonymous uploader",
	}
}

func (t Tag) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	err = e.EncodeElement(t.toFeed(), start)
	return
}
