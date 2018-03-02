package feed

import (
	"io"
)

// feed parser
type Parser interface {
	Decode(r io.Reader) ([]Torrent, error)
}
