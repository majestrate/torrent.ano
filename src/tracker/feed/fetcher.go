package feed

import (
	"net/http"
	"net/url"
	"sync"
	"time"
	"tracker/config"
	"tracker/db"
	"tracker/log"
)

type Fetch interface {
	Fetch() error
	Retry()
	Name() string
}

type feedEvent struct {
	client *http.Client
	f      *Fetcher
	conf   config.FeedConfig
	db     db.DB
	parser Parser
}

func (f *feedEvent) Name() string {
	return f.conf.Name
}

func (f *feedEvent) Fetch() (err error) {
	var torrents []Torrent
	log.Infof("fetching feed %s", f.Name())
	var resp *http.Response
	resp, err = f.client.Get(f.conf.URL)
	if err == nil {
		defer resp.Body.Close()
		if f.parser != nil {
			torrents, err = f.parser.Decode(resp.Body)
			if err == nil {
				for _, t := range torrents {
					var has bool
					has, err = f.db.HasTorrent(t.InfoHash)
					if err == nil && !has {
						f.f.QueueTorrent(t.URL.String())
					}
				}
			}
		}
	}
	return
}

func (f *feedEvent) Retry() {
	f.f.QueueFeed(f.conf)
}

type Fetcher struct {
	conf       config.FeedsConfig
	db         db.DB
	ticker     *time.Ticker
	pending    []Fetch
	pendingMtx sync.Mutex
	pool       chan []Fetch
	client     *http.Client
}

func (f *Fetcher) Run(numWorkers int) {
	f.pool = make(chan []Fetch)
	f.client = &http.Client{
		Transport: &http.Transport{
			Proxy: func(*http.Request) (*url.URL, error) {
				return f.conf.ProxyURL, nil
			},
		},
	}

	num := 1
	if numWorkers > 0 {
		num = numWorkers
	}

	for num > 0 {
		go f.worker()
		num--
	}

	for {
		f.pendingMtx.Lock()
		pending := f.pending
		f.pendingMtx.Unlock()

		if len(pending) > 0 {
			f.pool <- pending
		}

		f.pendingMtx.Lock()
		f.pending = nil
		f.pendingMtx.Unlock()
		<-f.ticker.C
	}
}

func (f *Fetcher) QueueTorrent(url string) {
	// TODO: implement
}

func (f *Fetcher) QueueFeed(conf config.FeedConfig) {
	f.pendingMtx.Lock()
	f.pending = append(f.pending, &feedEvent{
		client: f.client,
		conf:   conf,
		db:     f.db,
		f:      f,
	})
	f.pendingMtx.Unlock()
}

func (f *Fetcher) worker() {
	for {
		fetches := <-f.pool
		for _, fetch := range fetches {
			err := fetch.Fetch()
			if err != nil {
				log.Errorf("failed to fetch %s: %s", fetch.Name(), err)
				fetch.Retry()
			}
		}
	}
}

func NewFetcher(conf config.FeedsConfig, db db.DB) *Fetcher {
	return &Fetcher{
		conf:   conf,
		db:     db,
		ticker: time.NewTicker(time.Minute),
	}
}
