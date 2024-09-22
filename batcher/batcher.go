package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func NewBatcher(
	ctx context.Context,
	limit int,
	flushPeriod time.Duration,
	writer Writer,
) *Batcher {
	b := &Batcher{
		bufSize:     limit,
		flushPeriod: flushPeriod,

		ctx: ctx,

		cv:  sync.NewCond(&sync.RWMutex{}),
		buf: make([]string, 0, limit),

		ticker: time.NewTicker(flushPeriod),

		writer: writer,
	}

	go b.sync()

	return b
}

type (
	Writer interface {
		Write(context.Context, []interface{}) error
	}

	Batcher struct {
		bufSize     int
		flushPeriod time.Duration

		ctx context.Context

		cv  *sync.Cond
		buf []string

		ticker *time.Ticker

		err error

		writer Writer
	}
)

func (b *Batcher) Write(data string) error { // NB - do ctx needed?
	if len(b.buf) >= b.bufSize {
		b.flush(false)
	}

	b.cv.L.Lock()
	b.buf = append(b.buf, data)
	fmt.Printf("%s - wait\n", data)
	b.cv.Wait()
	defer b.cv.L.Unlock()
	// fmt.Printf("%s - return %v\n", data, b.err)
	return b.err
}

func (b *Batcher) flush(byTimer bool) {
	if byTimer {
		fmt.Printf("flush by timer\n")
	} else {
		fmt.Printf("flush by size\n")
	}

	b.cv.L.Lock()
	// defer b.cv.L.Unlock() // NB: is defer suitable

	batch := make([]interface{}, 0, len(b.buf))
	for i := range b.buf {
		batch = append(batch, b.buf[i])
	}

	b.buf = b.buf[:0] // error [:]

	b.cv.L.Unlock()

	b.err = b.writer.Write(b.ctx, batch)

	b.ticker.Reset(b.flushPeriod)

	b.cv.Broadcast()
}

func (b *Batcher) sync() {
	for {
		select {
		case <-b.ctx.Done():
			b.ticker.Stop()
		case <-b.ticker.C:
			b.flush(true)
		}
	}
}
