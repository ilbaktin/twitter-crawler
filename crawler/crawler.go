package main

import (
	"flag"
	"fmt"
	"github.com/scarecrow6977/twitter-crawler/crawler/conf"
	"github.com/scarecrow6977/twitter-crawler/crawler/crawler"
	"github.com/scarecrow6977/twitter-crawler/crawler/log"
	"github.com/scarecrow6977/twitter-crawler/crawler/storage/pg-storage"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "path to the config file")
	flag.Parse()

	conf.Init(configPath)
	config, err := conf.LoadConfig()
	if err != nil {
		fmt.Printf("can't load config, err='%v'", err)
		return
	}
	log.SetVerbosityLevel(2)

	pgStorage, err := pg_storage.NewPgStorage(config.PostgresAccess)
	if err != nil {
		log.LogError(fmt.Sprintf("can't load config, err='%v'", err))
		return
	}
	m := crawler.NewCrawlerMaster(config.NumOfWorkers, config.QueueSize, config.QueueNoRefillLimit, pgStorage)

	m.Run()
}
