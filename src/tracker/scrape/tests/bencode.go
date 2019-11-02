package main

import (
	"crypto/sha1"
	"fmt"
	b "github.com/felix/go-bencode"
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
func main() {
	file, err := os.Open("scrape")
	if err != nil {
		fmt.Println("Error with open file!")
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		fmt.Println("Error with get size of file")
		os.Exit(1)
	}
	fmt.Print("file size: ")
	fmt.Println(stat.Size())
	//	bs:=make([]byte, stat.Size())
	//	_, err = file.Read(bs)
	//	if err!=nil{
	//		fmt.Println("Error with read of file")
	//		os.Exit(1)
	//	}
	//	fmt.Println(bs)
	fmt.Println("decode init")
	var i interface{}
	packet := make([]byte, stat.Size())
	_, err = file.Read(packet)
	if err != nil {
		fmt.Println("Error with read file")
		os.Exit(1)
	}
	i, err = b.Decode(packet)
	if err != nil {
		fmt.Println("Error with decode")
		os.Exit(1)
	}
	tmp2 := FilesConstructMap(i)
	//tmp:=tmp0["files"].(map[string]map[interface{}]map[string]int)
	for k, v := range tmp2 {
		fmt.Print(k + ": ")
		fmt.Println(v)
	}
	//tmp:= i["files"].(map[string]interface{})

	m := map[string]map[string]string{
		"some": map[string]string{
			"some1": "value",
		},
	}
	fmt.Println(m)

	m1 := map[string]string{
		"Just string": "Too just string",
	}

	fmt.Println(m1)
}
