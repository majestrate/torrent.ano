package scrape

import "testing"
import config_scrape "./test_config_scrape"
import "fmt"


func TestDownloadFile(t *testing.T){
	err:= DownloadFile(config_scrape.DEFAULT_SCRAPE_FILE_PATH,config_scrape.DEFAULT_SCRAPE_URL);
	if err!=nil{
		t.Error("Err with download file")
	}

}

func TestReadScrape(t *testing.T){
	switch data:=ReadScrape(config_scrape.DEFAULT_SCRAPE_FILE_PATH).(type) {
		case map[string]map[string]string :

			fmt.Println("Its map");
			fmt.Println(data);
		default:
			fmt.Println("I get error");
			fmt.Println(data);

	}

}
