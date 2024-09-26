package main

import (
	"errors"
	"sync"
)

var (
	ErrFull   = errors.New("buffer full")
	ErrClosed = errors.New("shutting down")
)

type Pool struct {
	maxBufSize uint

	running bool

	m *sync.Mutex

	curBufSize uint
	jobs       chan Job

	wg *sync.WaitGroup
}

func (p *Pool) AddJob(j Job) error {
	p.m.Lock()
	defer p.m.Unlock()

	if !p.running {
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

	if !p.running {
		p.m.Unlock()
		return
	}

	p.running = false

	close(p.jobs)

	p.m.Unlock()

	p.wg.Wait()
}

func NewPool(workersCount uint) *Pool {
	p := Pool{
		maxBufSize: workersCount,
		running:    true,
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
