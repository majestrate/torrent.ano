package scrape

import "testing"
import config_scrape "./test_config_scrape"
import "fmt"

func TestDownloadFile(t *testing.T) {
	err := DownloadFile(config_scrape.DEFAULT_SCRAPE_FILE_PATH, config_scrape.DEFAULT_SCRAPE_URL)
	if err != nil {
		t.Error("Err with download file")
	}

}

func TestReadScrape(t *testing.T) {
	_, err := ReadScrape(config_scrape.DEFAULT_SCRAPE_FILE_PATH)
	if err != nil {
		t.Error(err)
	}
}

func TestGetScrapeByInfoHash(t *testing.T) {
	err, mp := GetScrapeByInfoHash(config_scrape.DEFAULT_SCRAPE_FILE_PATH, config_scrape.DEFAULT_SCRAPE_URL, "9865c87afbdf138aa3e2b88220187f38015dcfc0")
	if err != nil {
		t.Error(err)
	}
	for key, value := range mp {
		fmt.Print(key + ": ")
		fmt.Println(value)
	}
}
