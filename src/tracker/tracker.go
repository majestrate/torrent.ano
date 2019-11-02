package tracker

import (
	"net/http"
	"os"
	"tracker/config"
	"tracker/db"
	"tracker/feed"
	"tracker/index"
	"tracker/log"
)

func Run() {
	fname := "default.ini"
	if len(os.Args) > 1 {
		fname = os.Args[1]
	}
	log.SetLevel("info")
	cfg := new(config.Config)
	err := cfg.Load(fname)
	if err != nil {
		log.Fatalf("%s", err)
	}

	log.SetLevel(cfg.Log.Level)

	idx := index.New(&cfg.Index)
	idx.DB, err = db.NewPostgres(&cfg.DB)
	//idx.Cfg_scrape.Load(&cfg.Scrape) // TODO
	if err != nil {
		log.Fatalf("%s", err)
	}
	err = idx.DB.Init()
	if err != nil {
		log.Fatalf("%s", err)
	}
	if cfg.Feeds.Enabled {
		fetcher := feed.NewFetcher(cfg.Feeds, idx.DB)
		go fetcher.Run(cfg.Feeds.Jobs)
	}
	addr := cfg.Index.Addr
	log.Infof("serve http at http://%s/", addr)
	err = http.ListenAndServe(addr, idx)
	if err != nil {
		log.Fatalf("http serve failed: %s", err)
	}
}
