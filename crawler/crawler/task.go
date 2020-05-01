package crawler

import "github.com/scarecrow6977/twitter-crawler/crawler/storage"

type CrawlerTask interface {
	Exec(stor storage.Storage) error
}
