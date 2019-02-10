package store

/// Article query info
type ArticleInfo struct {
	Group     string
	Number    uint64
	MessageID string
}

type ArticleRange struct {
	Group string
	Hi    uint64
	Lo    uint64
}
