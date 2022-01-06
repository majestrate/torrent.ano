package config

import (
	"github.com/majestrate/torrent.ano/lib/tracker/config/parser"
)

type DBConfig struct {
	Type string
	URL  string
}

func (cfg *DBConfig) Load(s *parser.Section) error {
	cfg.Type = s.Get("type", "postgres")
	cfg.URL = s.Get("url", "")
	return nil
}
