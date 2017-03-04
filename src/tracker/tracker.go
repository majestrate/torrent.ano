package main

import (
	"net/http"
	"os"
	"tracker/config"
	"tracker/db"
	"tracker/index"
	"tracker/log"
)

func main() {
	fname := "default.ini"
	if len(os.Args) > 1 {
		fname = os.Args[1]
	}
	log.SetLevel("debug")
	cfg := new(config.Config)
	err := cfg.Load(fname)
	if err != nil {
		log.Fatalf("%s", err)
	}
	idx := index.New(&cfg.Index)
	idx.DB, err = db.NewPostgres(&cfg.DB)
	if err != nil {
		log.Fatalf("%s", err)
	}
	err = idx.DB.Init()
	if err != nil {
		log.Fatalf("%s", err)
	}
	addr := cfg.Index.Addr
	log.Infof("serve http at http://%s/", addr)
	err = http.ListenAndServe(addr, idx)
	if err != nil {
		log.Fatalf("http serve failed: %s", err)
	}
}
