package crawler

import "univer/twitter-crawler/storage"

type CrawlerTask interface {
	Exec(stor storage.Storage) error
}
