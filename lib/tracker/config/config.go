package config

import (
	"github.com/majestrate/torrent.ano/lib/tracker/config/parser"
)

type configLoadable interface {
	Load(s *parser.Section) error
}

type Config struct {
	DB     DBConfig
	Index  IndexConfig
	Log    LogConfig
	Feeds  FeedsConfig
	Scrape ScrapeConfig
}

func (cfg *Config) Load(fname string) error {
	sections := map[string]configLoadable{
		"db":     &cfg.DB,
		"index":  &cfg.Index,
		"log":    &cfg.Log,
		"feeds":  &cfg.Feeds,
		"scrape": &cfg.Scrape,
	}

	conf, err := parser.Read(fname)
	if err != nil {
		return err
	}
	for sect := range sections {
		s, err := conf.Section(sect)
		if err != nil {
			return err
		}
		err = sections[sect].Load(s)
		if err != nil {
			return err
		}
	}
	return nil
}
