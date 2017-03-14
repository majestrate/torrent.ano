package main

import (
	"fmt"
	"os"
	"strings"
	"tracker/config"
	"tracker/db"
	"tracker/log"
)

const AddUser = "add-user"
const DelUser = "del-user"

func printUsage() {
	fmt.Fprintf(os.Stdout, "usage: %s config.ini [%s username password|%s username] ...\n", os.Args[0], AddUser, DelUser)
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
	} else {
		printUsage()
	}
}
