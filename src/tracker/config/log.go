package config

import (
	"tracker/config/parser"
)

type LogConfig struct {
	Level string
}

func (cfg *LogConfig) Load(s *parser.Section) (err error) {
	cfg.Level = s.Get("level", "info")
	return
}
