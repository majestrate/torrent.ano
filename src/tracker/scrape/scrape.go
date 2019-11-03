package scrape

import (
	"fmt"
	b "github.com/felix/go-bencode"

	"crypto/sha1"
	"io"
	"os"
)

type Files struct{
	hash string
	downloaded, complete, incomplete int64
};

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
			switch K{
				case "complete":
					complete=V.(int64)
				case "incomplete":
					incomplete=V.(int64)
				case "downloaded":
					downloaded=V.(int64)
			}
		}
		newItem := Files{ 
			downloaded: downloaded,
			complete: complete,
			incomplete: incomplete,
		}
		ret = append(ret, newItem)
	}
	return ret
}

func ReadScrape(file_path string) (interface{}, error) {
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
	packet := make([]byte, stat.Size())
	_, err = file.Read(packet)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	data, err := b.Decode(packet)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return data, nil
}
