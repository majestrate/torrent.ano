package scrape

import (
	"fmt"
	b "github.com/felix/go-bencode"

	"crypto/sha1"
	"io"
	"os"
)

func FilesConstructMap(raw interface{}) map[string]map[string]int64 {
	tmp := raw.(map[string]interface{})
	tmp1 := tmp["files"]
	tmp2 := tmp1.(map[string]interface{})
	ret := make(map[string]map[string]int64)

	h := sha1.New()
	for key, value := range tmp2 {
		io.WriteString(h, key)
		key = fmt.Sprintf("%x", h.Sum(nil))
		tmp_ := value.(map[string]interface{})

		ret[key] = make(map[string]int64)
		for K, V := range tmp_ {
			ret[key][K] = V.(int64)
		}
	}
	return ret
}

func ReadScrape(file_path string) interface{} {
	file, err := os.Open(file_path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()
	//check if file is not null
	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return err
	}
	if stat.Size() == 0 {
		fmt.Println("Scrape file is null size")
		return err
	}
	var tmp0 interface{}
	tmp, err := b.Decode(file)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return data
}
