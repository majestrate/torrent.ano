package config

import (
	"tracker/config/parser"
)

type configLoadable interface {
	Load(s *parser.Section) error
}

type Config struct {
	DB    DBConfig
	Index IndexConfig
	Log   LogConfig
}

func (cfg *Config) Load(fname string) error {
	sections := map[string]configLoadable{
		"db":    &cfg.DB,
		"index": &cfg.Index,
		"log":   &cfg.Log,
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
