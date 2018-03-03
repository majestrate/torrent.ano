package config

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"tracker/config/parser"
	"tracker/log"
)

const DefaultFeedsDir = "feeds.d"
const DefaultFeedsFile = "feeds.ini"
const DefaultFeedsProxyURL = "http://127.0.0.1:8118/"
const DefaultFeedsJobs = 4

type FeedConfig struct {
	Name       string
	URL        string
	CategoryID uint64
}

type FeedsConfig struct {
	Directory string
	File      string
	ProxyURL  *url.URL
	Jobs      int
	Feeds     []FeedConfig
	Enabled   bool
}

func (cfg *FeedsConfig) Load(s *parser.Section) (err error) {
	cfg.Enabled = s.Get("enabled", "1") == "1"
	if cfg.Enabled {
		cfg.Directory = s.Get("dir", DefaultFeedsDir)
		cfg.File = s.Get("file", DefaultFeedsFile)
		cfg.ProxyURL, err = url.Parse(s.Get("proxy", DefaultFeedsProxyURL))
		cfg.Jobs = s.GetInt("jobs", DefaultFeedsJobs)
		if err == nil {
			var feedSections []*parser.Section
			var c *parser.Configuration
			var files []string
			var infos []os.FileInfo
			infos, err = ioutil.ReadDir(cfg.Directory)
			if err == nil {
				for _, info := range infos {
					f := info.Name()
					if strings.HasSuffix(f, ".ini") {
						files = append(files, filepath.Join(cfg.Directory, f))
					}
				}
			}
			files = append(files, cfg.File)
			for _, f := range files {
				c, err = parser.Read(f)
				if err == nil {
					var sects []*parser.Section
					sects, err = c.AllSections()
					if err == nil {
						feedSections = append(feedSections, sects...)
					}
				}
			}

			for _, sect := range feedSections {
				url := sect.Get("url", "")
				catid := sect.GetInt("category", 0)
				if url != "" && catid != 0 {
					name := sect.Name()
					log.Infof("loaded feed %s", name)
					cfg.Feeds = append(cfg.Feeds, FeedConfig{
						Name:       name,
						URL:        url,
						CategoryID: uint64(catid),
					})
				}
			}
		}
	}
	return
}
