package feed

import (
	"net/url"
)

type Torrent struct {
	URL      *url.URL
	InfoHash [20]byte
}
