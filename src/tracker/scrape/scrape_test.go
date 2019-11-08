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
	fmt.Printf( "D: %d; C: %d, I:%d",mp.Downloaded, mp.Complete, mp.Incomplete )
}

func TestGetScrapeByInfoHashInTrackers(t * testing.T){
	trackers:= []string{"http://tracker.postman.i2p/announce"}

	err, mp := GetScrapeByInfoHashInTrackers(config_scrape.DEFAULT_SCRAPE_FILE_PATH, trackers, "9865c87afbdf138aa3e2b88220187f38015dcfc0", "http://127.0.0.1:4444" )
	if err != nil{
			t.Error(err)
	}
	for m := range mp{
			fmt.Printf( "D: %d; C: %d, I:%d",mp[m].Downloaded, mp[m].Complete, mp[m].Incomplete )
	}

}
