package main

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Job = func()

var ErrNoSuchJob = errors.New("no such job")

func NewSimpleScheduler() *SimpleScheduler {
	return &SimpleScheduler{
		m:    sync.RWMutex{},
		jobs: make(map[uuid.UUID]chan struct{}),
	}
}

type SimpleScheduler struct {
	m    sync.RWMutex
	jobs map[uuid.UUID]chan struct{}
}

func (s *SimpleScheduler) SetTimeout(d time.Duration, j Job) uuid.UUID {
	id := uuid.New()

	cancel := make(chan struct{})

	done := make(chan struct{})

	go func() {
		t := time.NewTimer(d)

		select {
		case <-t.C:
			j()
		case <-cancel:
			t.Stop()
		}

		close(done)
	}()

	go func() {
		<-done

		s.m.Lock()
		defer s.m.Unlock()

		delete(s.jobs, id)
	}()

	s.m.Lock()
	defer s.m.Unlock()

	// todo: what if id is already exists in s.jobs as key?
	s.jobs[id] = cancel

	return id
}

func (s *SimpleScheduler) CancelTimeout(id uuid.UUID) error {
	s.m.Lock()
	defer s.m.Unlock()

	cancel, ok := s.jobs[id]
	if !ok {
		return ErrNoSuchJob
	}

	close(cancel)

	delete(s.jobs, id)

	return nil
}

func (s *SimpleScheduler) Stop() {
	// TODO
}
