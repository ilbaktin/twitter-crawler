package crawler

import (
	"github.com/hako/durafmt"
	"github.com/pkg/errors"
	crawler_tasks "github.com/scarecrow6977/twitter-crawler/crawler/crawler-tasks"
	"github.com/scarecrow6977/twitter-crawler/crawler/log"
	"github.com/scarecrow6977/twitter-crawler/crawler/storage"
	"sync"
	"time"
)

type CrawlerMaster struct {
	workers   []*CrawlerWorker
	taskQueue chan CrawlerTask
	lock      sync.Mutex

	stor         storage.Storage
	workerDoneCh chan int

	queueSize          int
	queueNoRefillLimit int

	*log.Logger
}

func NewCrawlerMaster(numOfWorkers int, queueSize int, queueNoRefillLimit int, stor storage.Storage) *CrawlerMaster {
	logger := log.NewLogger("CrawlerMaster")

	master := &CrawlerMaster{
		taskQueue:          make(chan CrawlerTask, queueSize),
		workerDoneCh:       make(chan int),
		stor:               stor,
		queueSize:          queueSize,
		queueNoRefillLimit: queueNoRefillLimit,
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
		case <-time.After(5 * time.Second):

			currentTime := time.Now()
			timePassed := durafmt.Parse(currentTime.Sub(startTime)).LimitFirstN(3).String()
			m.LogInfo("Time passed: %s", timePassed)

			if len(m.taskQueue) < m.queueNoRefillLimit {
				newTasksAmount, err := m.refillTaskQueue()
				if err != nil {
					return errors.Wrap(err, "get users with not downloaded followers")
				}
				m.LogInfo("%d new tasks received", newTasksAmount)
			}
		}
	}
	return nil
}

func (m *CrawlerMaster) refillTaskQueue() (int64, error) {
	// Этот код не универсальный, заточен под единственный вид тасков, TODO: придумать как обобщить
	newTasksAmount := int64(m.queueSize - m.queueNoRefillLimit)
	users, err := m.stor.GetUsersWithNotDownloadedFollowers(newTasksAmount)
	if err != nil {
		return 0, err
	}
	for _, user := range users {
		m.taskQueue <- &crawler_tasks.DownloadFollowersTask{
			ScreenName: user.ScreenName,
		}
	}
	return newTasksAmount, nil
}
