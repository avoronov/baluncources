package main

import (
	"fmt"
	"time"
)

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
