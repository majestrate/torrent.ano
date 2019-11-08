package scrape

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	net_url "net/url"
	"os"
	"errors"
)

func DownloadFile(filepath string, url string, proxies ... string) (err error) {
	var response *http.Response
	if len(proxies) > 0{
		for proxy := range proxies{
			proxyURL, err := net_url.Parse( proxies[proxy] )
			if err != nil{
				err = err
				continue
			}
			urlParsed, err := net_url.Parse( url )
			if err != nil{
				err = err
				continue
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
			client := &http.Client{
				Transport: transport,
			}
			request, err := http.NewRequest("GET", urlParsed.String(), nil)
			if err != nil{
				err = err
				continue
			}
			response, err = client.Do(request)
			if err != nil{
				err = err
				continue
			}
			break;
		}
		if err != nil{
			return err
		}
	}else{
		response, err = http.Get(url)
		if err != nil {
			return err
		}
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
func GetScrapeByInfoHash(filepath string, url string, hash string, proxies ... string) (err error, mp Files) {
	var n int

	src := []byte(hash)
	bin := make([]byte, hex.DecodedLen(len(src)))

	n, err = hex.Decode(bin, src)
	if err != nil {
		return err, mp
	}
	info_hash := fmt.Sprintf("%s", net_url.QueryEscape(string(bin[:n])))
	FileURL, err := net_url.Parse( url+"?info_hash="+info_hash );
	if err != nil{
		return err, mp
	}
	filepathFull:=filepath+"_"+hash

	if err := DownloadFile( filepathFull , FileURL.String(), proxies... ); err != nil {
		return err, mp
	}
	raw, err := ReadScrape(filepathFull)
	if err != nil {
		return err, mp
	}
	mp = FilesConstructMap(raw)[0]
	return nil, mp
}

func GetScrapeByInfoHashInTrackers(filepath string, trackers []string, hash string, proxy string) (error, []Files){
	var mp []Files
	for tracker := range trackers{
		err, r:=GetScrapeByInfoHash(filepath, trackers[tracker], hash, proxy)
		if err != nil{
			continue
		}
		mp=append(mp, r)
	}
	if len(mp) == 0{
		return errors.New("Can't get scrape thought proxy"), nil
	}
	return nil, mp
}



