package scrape

import(
	"net/http"
	"io"
	"os"
	config_scrape "./config_scrape"
)


func DownloadFile(filepath string, url string){
	response, err = http.Get(url);
	if err!=nil{ return err	}
	defer response.Body.Close();

	outfile, err = os.Create(config_scrape.DEFAULT_SCRAPE_FILE_PATH);
	if err!=nil{ return err	}
	defer outfile.Close();

	_,err =io.Copy(outfile.response.Body)
	return err;
}
