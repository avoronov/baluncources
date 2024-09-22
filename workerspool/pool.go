package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrFull = errors.New("buffer full")
var ErrClosed = errors.New("shutting down")

type Pool struct {
	maxBufSize uint

	runnig bool

	m *sync.Mutex

	curBufSize uint
	jobs       chan Job

	wg *sync.WaitGroup
}

func (p *Pool) AddJob(j Job) error {
	p.m.Lock()
	defer p.m.Unlock()

	if !p.runnig {
		return ErrClosed
	}

	if p.curBufSize < p.maxBufSize {
		p.curBufSize++
		p.jobs <- j
	} else {
		return ErrFull
	}

	return nil
}

func (p *Pool) Close() {
	p.m.Lock()

	if !p.runnig {
		p.m.Unlock()
		return
	}

	p.runnig = false

	close(p.jobs)

	p.m.Unlock()

	p.wg.Wait()
}

func NewPool(workersCount uint) *Pool {
	p := Pool{
		maxBufSize: workersCount,
		runnig:     true,
		m:          &sync.Mutex{},
		jobs:       make(chan Job, workersCount),
		wg:         &sync.WaitGroup{},
	}

	p.wg.Add(int(workersCount))

	for i := 0; i < int(workersCount); i++ {
		go p.worker()
	}

	return &p
}

func (p *Pool) worker() {
	defer p.wg.Done()

	for j := range p.jobs {
		p.m.Lock()
		p.curBufSize--
		p.m.Unlock()

		j.Do()
	}
}

type Job struct {
	id     uint32
	period time.Duration
}

func (j Job) Do() {
	time.Sleep(j.period)

	fmt.Printf("job %d is done\n", j.id)
}

func NewJob(id, p uint32) Job {
	return Job{
		id:     id,
		period: time.Duration(p) * time.Millisecond,
	}
}
