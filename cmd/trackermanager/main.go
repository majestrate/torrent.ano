package main

import (
	"fmt"
	"github.com/majestrate/torrent.ano/lib/tracker/config"
	"github.com/majestrate/torrent.ano/lib/tracker/db"
	"github.com/majestrate/torrent.ano/lib/tracker/log"
	"os"
	"strings"
)

const AddUser = "add-user"
const DelUser = "del-user"
const AddCategory = "add-category"
const DelCategory = "del-category"
const DelTorrent = "del-torrent"

func printUsage() {
	fmt.Fprintf(os.Stdout, "usage: %s config.ini [%s username password|%s username|%s name|%s name|%s infohash] ...\n", os.Args[0], AddUser, DelUser, AddCategory, DelCategory, DelTorrent)
}

func main() {

	if len(os.Args) < 3 {
		printUsage()
		return
	}

	fname := os.Args[1]
	action := strings.ToLower(os.Args[2])

	cfg := new(config.Config)
	err := cfg.Load(fname)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.SetLevel(cfg.Log.Level)
	DB, err := db.NewPostgres(&cfg.DB)
	if action == AddUser {
		if len(os.Args) < 4 {
			log.Error("not enough arguments")
			return
		}
		username := os.Args[3]
		password := os.Args[4]
		err = DB.AddUserLogin(username, password)
		if err != nil {
			log.Errorf("didnt add new user %s: %s", username, err)
			return
		}
		log.Infof("added user %s", username)
	} else if action == DelUser {
		username := os.Args[3]
		err = DB.DelUserLogin(username)
		if err != nil {
			log.Errorf("didnt remove user %s: %s", username, err)
		}
	} else if action == AddCategory {
		err = DB.AddCategory(os.Args[3])
	} else if action == DelCategory {
		err = DB.DelCategory(os.Args[3])
	} else if action == DelTorrent {
		err = DB.DelTorrent(os.Args[3])
	} else {
		printUsage()
		return
	}
	if err != nil {
		log.Errorf("error: %s", err.Error())
	}
}
