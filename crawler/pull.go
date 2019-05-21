package crawler

import (
	"sync"
)

type TaskPull struct {
	data	[]CrawlerTask
	lock	sync.Mutex
}

func NewTaskPull(tasks []CrawlerTask) *TaskPull {
	return &TaskPull{
		data: tasks,
	}
}

func (p *TaskPull) IsEmpty() bool {
	return len(p.data) == 0
}

func (p *TaskPull) Get() CrawlerTask {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.IsEmpty() {
		return nil
	}

	task := p.data[0]
	p.data = p.data[1:]

	return task
}


