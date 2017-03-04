package config

import (
	"net/url"
	"tracker/config/parser"
)

const DefaultTrackerURL = "http://21.3.37.31:6881/announce"

type IndexConfig struct {
	CaptchaWidth  int
	CaptchaHeight int
	TrackerURL    *url.URL
	Addr          string
	TemplateDir   string
	StaticDir     string
	TorrentsDir   string
}

func (cfg *IndexConfig) Load(s *parser.Section) (err error) {
	cfg.CaptchaWidth = s.GetInt("captcha-width", 400)
	cfg.CaptchaHeight = s.GetInt("captcha-height", 100)
	cfg.TrackerURL, err = url.Parse(s.Get("tracker-url", DefaultTrackerURL))
	cfg.Addr = s.Get("bind", "[::]:1800")
	cfg.TemplateDir = s.Get("template-dir", "./templates/")
	cfg.StaticDir = s.Get("static-dir", "./static/")
	cfg.TorrentsDir = s.Get("torrents-dir", "./torrents/")
	return nil
}
