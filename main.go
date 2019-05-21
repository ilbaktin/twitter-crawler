package main

import (
	"fmt"
	"univer/twitter-crawler/conf"
	"univer/twitter-crawler/crawler"
	"univer/twitter-crawler/crawler-tasks"
	"univer/twitter-crawler/log"
	"univer/twitter-crawler/storage/pg-storage"
)

func main() {
	conf.Init("config.yaml")
	config, err := conf.LoadConfig()
	if err != nil {
		fmt.Printf("can't load config, err='%v'", err)
		return
	}
	log.SetVerbosityLevel(2)

	tasks := []crawler.CrawlerTask{
		crawler_tasks.DownloadFollowersTask{
			ScreenName: "ZBaHJl5FPXtwmtF",
		},
	}
	pull := crawler.NewTaskPull(tasks)

	pgStorage, err := pg_storage.NewPgStorage(config.PostgresAccess)
	if err != nil {
		log.LogError(fmt.Sprintf("can't load config, err='%v'", err))
		return
	}
	m := crawler.NewCrawlerMaster(config.NumOfWorkers, pull, pgStorage)

	m.Run()
}
