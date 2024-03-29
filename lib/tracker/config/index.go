package config

import (
	"github.com/majestrate/torrent.ano/lib/tracker/config/parser"
	"net/url"
)

const DefaultTrackerURL = "http://tracker.livingstone.i2p/a"
const DefaultSiteName = "TORRENTS.LIVINGSTONE.I2P"

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
	return
}
