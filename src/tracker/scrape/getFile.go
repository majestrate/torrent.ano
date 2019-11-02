package scrape

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	net_url "net/url"
	"os"
)

func DownloadFile(filepath string, url string) (err error) {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Println(response.StatusCode)
	}
	outfile, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer outfile.Close()
	_, err = io.Copy(outfile, response.Body)
	return err
}

//hash from psql
//url to scrape
func GetScrapeByInfoHash(filepath string, url string, hash string) (err error, mp map[string]map[string]int64) {
	var n int

	src := []byte(hash)
	bin := make([]byte, hex.DecodedLen(len(src)))

	n, err = hex.Decode(bin, src)
	if err != nil {
		return err, nil
	}
	info_hash := fmt.Sprintf("%s", net_url.QueryEscape(string(bin[:n])))
	if err := DownloadFile(filepath+"_"+hash, url+"?info_hash="+info_hash); err != nil {
		return err, nil
	}
	raw, err := ReadScrape(filepath + "_" + hash)
	if err != nil {
		return err, nil
	}
	mp = FilesConstructMap(raw)
	return

}
