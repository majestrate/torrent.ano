package config

import (
	"github.com/majestrate/torrent.ano/lib/tracker/config/parser"
)

type NNTPConfig struct {
	Addr    string
	Enabled bool
}

func (cfg *NNTPConfig) Load(s *parser.Section) (err error) {
	cfg.Addr = s.Get("addr", "127.0.0.1:1119")
	cfg.Enabled = s.Get("enabled", "0") == "1"
	return
}
