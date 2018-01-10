package config

import (
	"net/url"
	"tracker/config/parser"
)

const DefaultTrackerURL = "http://uajd4nctepxpac4c4bdyrdw7qvja2a5u3x25otfhkptcjgd53ioq.b32.i2p/announce"
const DefaultSiteName = "torrent.ano"

type IndexConfig struct {
	CaptchaWidth  int
	CaptchaHeight int
	TrackerURL    *url.URL
	Addr          string
	TemplateDir   string
	StaticDir     string
	TorrentsDir   string
	SiteName      string
}

func (cfg *IndexConfig) Load(s *parser.Section) (err error) {
	cfg.CaptchaWidth = s.GetInt("captcha-width", 400)
	cfg.CaptchaHeight = s.GetInt("captcha-height", 100)
	cfg.TrackerURL, err = url.Parse(s.Get("tracker-url", DefaultTrackerURL))
	cfg.Addr = s.Get("bind", "[::]:1800")
	cfg.TemplateDir = s.Get("template-dir", "./templates/")
	cfg.StaticDir = s.Get("static-dir", "./static/")
	cfg.TorrentsDir = s.Get("torrents-dir", "./torrents/")
	cfg.SiteName = s.Get("site-name", DefaultSiteName)
	return nil
}
