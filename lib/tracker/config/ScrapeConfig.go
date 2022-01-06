package config

import (
	"github.com/majestrate/torrent.ano/lib/tracker/config/parser"
)

const DEFAULT_SCRAPE_FILE_PATH string = "/tmp/scrape_torrent_ano"
const DEFAULT_SCRAPE_URL string = "http://127.0.0.1:7662/scrape"
const DEFAULT_ENABLED bool = true
const DEFAULT_ENABLED_STRING string = "true"

type ScrapeConfig struct {
	URL       string
	Enabled   bool
	File_path string
}

func (cfg *ScrapeConfig) Load(s *parser.Section) (err error) {
	cfg.URL = s.Get("url", DEFAULT_SCRAPE_URL)
	switch v := s.Get("enabled", DEFAULT_ENABLED_STRING); v {
	case "true":
		cfg.Enabled = true
	case "false":
		cfg.Enabled = true
	default:
		cfg.Enabled = DEFAULT_ENABLED
	}
	cfg.File_path = s.Get("file_path", DEFAULT_SCRAPE_FILE_PATH)
	return
}
