package crawler

import (
	"fmt"
	"sync"
	"time"
	"univer/twitter-crawler/log"
	"univer/twitter-crawler/storage"
)

type CrawlerMaster struct {
	workers			[]*CrawlerWorker
	taskPull		*TaskPull
	lock			sync.Mutex

	stor			storage.Storage
	workerDoneCh	chan int

	*log.Logger
}

func NewCrawlerMaster(numOfWorkers int, pull *TaskPull, stor storage.Storage) *CrawlerMaster {
	logger := log.NewLogger("CrawlerMaster")

	master := &CrawlerMaster{
		taskPull: pull,
		workerDoneCh: make(chan int),
		stor: stor,
	}
	master.Logger = logger
	workers := make([]*CrawlerWorker, 0, numOfWorkers)
	for i := 0; i < numOfWorkers; i++ {
		workers = append(workers, NewCrawlerWorker(i, master))
	}
	master.workers = workers

	return master
}

func (m *CrawlerMaster) Run() error {
	m.LogInfo("Starting crawler master...")
	for _, w := range m.workers {
		go w.Run()
	}

	startTime := time.Now()
	workersDone := 0

	loop:
	for {
		select {
		case _ = <-m.workerDoneCh:
			workersDone++
			if workersDone == len(m.workers) {
				break loop
			}
		case <-time.After(10 * time.Second):
			currentTime := time.Now()
			timePassed := currentTime.Sub(startTime)
			m.LogInfo(fmt.Sprintf("Time passed: %v", timePassed))
			m.LogInfo("Waiting for workers...")
		}
	}
	return nil
}
