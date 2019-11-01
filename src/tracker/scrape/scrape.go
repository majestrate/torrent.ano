package scrape;

import (
	b "github.com/jackpal/bencode-go"
	"fmt"
        "os"
	config_scrape "./config_scrape"
)


func ReadScrape() (interface{}){
	file,err:=os.Open(config_scrape.DEFAULT_SCRAPE_FILE_PATH)
	if err != nil{
		fmt.Println(err)
		return err
	}
	defer file.Close();
	//check if file is not null
	stat, err := file.Stat();
	if err !=nil{
		fmt.Println(err);
		return err
	}
	if(stat.Size() == 0){;
		fmt.Println("Scrape file is null size");
		return err
	}
	data, err := b.Decode(file)
	if err!=nil{
		fmt.Println(err);
		return err
	}
	return data

}
