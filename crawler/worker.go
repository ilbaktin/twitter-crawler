package crawler

import (
	"fmt"
	"univer/twitter-crawler/log"
)

type CrawlerWorker struct {
	id      int
	master  *CrawlerMaster
	isAlive bool
	*log.Logger
}

func NewCrawlerWorker(id int, m *CrawlerMaster) *CrawlerWorker {
	logger := log.NewLogger(fmt.Sprintf("CrawlerWorker %d", id))

	worker := &CrawlerWorker{
		id:     id,
		master: m,
	}
	worker.Logger = logger
	return worker
}

func (w *CrawlerWorker) IsAlive() bool {
	return w.isAlive
}

func (w *CrawlerWorker) Run() {
	w.LogInfo("Starting worker")

	w.isAlive = true
	for w.isAlive {
		task, ok := <-w.master.taskQueue
		if !ok {
			w.master.workerDoneCh <- w.id
			break
		}
		err := task.Exec(w.master.stor)
		if err != nil {
			w.LogError("error running task, err='%v'", err)
			//w.master.workerDoneCh <- w.id
			//break
			continue
		}
	}
	w.isAlive = false
	w.LogInfo("Exited")
}

func (w *CrawlerWorker) Kill() {
	w.LogInfo("Stopping worker...")
	w.isAlive = false
}
