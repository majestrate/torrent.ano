package model

import (
	"encoding/xml"
	"net/url"
	"time"
)

type FeedEntry interface {
	xml.Marshaler
	UploadedAt() time.Time
}

type AtomFeed struct {
	Domain  string
	BaseURL *url.URL
	Entries []FeedEntry
	Title   string
	ID      string
}

type Link struct {
	URL string `xml:"href,attr"`
}

func NewQueryLink(domain, path, query string) Link {
	u := &url.URL{
		Scheme:   "http",
		Host:     domain,
		Path:     path,
		RawQuery: query,
	}
	return Link{
		URL: u.String(),
	}
}
func NewLink(domain, path, fragment string) Link {
	u := &url.URL{
		Scheme:   "http",
		Host:     domain,
		Path:     path,
		Fragment: fragment,
	}
	return Link{
		URL: u.String(),
	}
}

type atomFeedImpl struct {
	Title    string      `xml:"title"`
	SubTitle string      `xml:"subtitle"`
	ID       string      `xml:"id"`
	Link     Link        `xml:"link"`
	Updated  time.Time   `xml:"updated"`
	Entries  []FeedEntry `xml:"entry"`
}

func (feed *AtomFeed) toFeed() *atomFeedImpl {

	latest := time.Unix(0, 0)
	for _, ent := range feed.Entries {
		u := ent.UploadedAt()
		if u.After(latest) {
			latest = u
		}
	}

	u := feed.BaseURL

	return &atomFeedImpl{
		Title:    feed.Title,
		SubTitle: feed.Title,
		Link:     NewLink(feed.Domain, u.RequestURI(), ""),
		ID:       feed.ID,
		Entries:  feed.Entries,
		Updated:  latest,
	}
}

func (feed *AtomFeed) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	start.Name.Local = "feed"
	start.Name.Space = "http://www.w3.org/2005/Atom"
	err = e.EncodeElement(feed.toFeed(), start)
	return
}
