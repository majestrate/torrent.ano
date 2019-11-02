package scrape

import (
	"io"
	"net/http"
	"os"
)

func DownloadFile(filepath string, url string) (err error) {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	outfile, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer outfile.Close()

	_, err = io.Copy(outfile, response.Body)
	return err
}
