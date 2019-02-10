package store

import "io"

// A write only message store
type WriteStore interface {
	// put article from io.reader into the store
	PutArticle(msg io.Reader) error
}

// A read only message store
type ReadStore interface {
	/// return nil. true if we have an article given the info
	HasArticle(info *ArticleInfo) (error, bool)
	/// read single article given info into visitor
	VisitArticle(info *ArticleInfo, v Visitor) error

	/// read range of articles into visitor
	VisitRange(r *ArticleRange, v Visitor) error
}

type Store interface {
	ReadStore
	// initialize store on disk
	// ensure all files and metadata
	Init() error
}
