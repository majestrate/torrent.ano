package store

import (
	"mime/multipart"
)

type Visitor interface {
	/// called on beginning
	BeginMessage(info *ArticleInfo) error
	// called for each mime header key/value
	Header(k, v string) error
	// called for each mime part or on a plaintext part if not multipart
	Part(p *multipart.Part) error
	// we ended a visit to a single message
	EndMessage(info *ArticleInfo, err error) error
	// called on visit end and propagate error if one occured
	End(err error)
}
