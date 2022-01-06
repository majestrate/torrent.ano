package scrape

import (
	"fmt"
	"github.com/zeebo/bencode"

	"crypto/sha1"
	"io"
	"os"
)

type Files struct {
	Hash                             string
	Downloaded, Complete, Incomplete int64
}

func FilesConstructMap(raw interface{}) []Files {
	tmp := raw.(map[string]interface{})
	tmp1 := tmp["files"]
	tmp2 := tmp1.(map[string]interface{})
	ret := make([]Files, 0)

	h := sha1.New()
	for key, value := range tmp2 {
		io.WriteString(h, key)
		key = fmt.Sprintf("%x", h.Sum(nil))
		tmp_ := value.(map[string]interface{})

		var downloaded, complete, incomplete int64
		for K, V := range tmp_ {
			switch K {
			case "complete":
				complete = V.(int64)
			case "incomplete":
				incomplete = V.(int64)
			case "downloaded":
				downloaded = V.(int64)
			}
		}
		newItem := Files{
			Downloaded: downloaded,
			Complete:   complete,
			Incomplete: incomplete,
		}
		ret = append(ret, newItem)
	}
	return ret
}

func ReadScrape(file_path string) (ret interface{}, err error) {
	file, err := os.Open(file_path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()
	//check if file is not null
	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if stat.Size() == 0 {
		fmt.Println("Scrape file is null size")
		return nil, err
	}
	dec := bencode.NewDecoder(file)
	err = dec.Decode(&ret)
	return
}
