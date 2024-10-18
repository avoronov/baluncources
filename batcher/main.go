package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func main() {
	const (
		workerCounts = 11
		tries        = 5
		bufLimit     = 10
		flushPeriod  = 500
	)

	var (
		ctx, cancel = context.WithCancel(context.Background())

		wg = sync.WaitGroup{}

		b = NewBatcher(ctx, bufLimit, time.Duration(flushPeriod*time.Millisecond), NewDummyWriter())
	)

	wg.Add(workerCounts)

	for i := range workerCounts {
		go func(i int) {
			for j := range tries {
				payload := fmt.Sprintf("'worker %d, try %d'", i, j)

				if err := b.Write(payload); err != nil {
					fmt.Printf("worker %d, try %d - error %s\n", i, j, err)

					continue
				}

				fmt.Printf("worker %d, try %d - success\n", i, j)

				// time.Sleep(1 * time.Second)
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

	cancel()
}

const failCount = 3

var ErrWrite = errors.New("write error")

func NewDummyWriter() *DummyWriter {
	return &DummyWriter{}
}

type DummyWriter struct {
	count uint
}

func (w *DummyWriter) Write(_ context.Context, data []interface{}) error {
	w.count++

	if (w.count % 10) == failCount {
		fmt.Printf("write batch %+v - error\n", data)
		return ErrWrite
	}

	fmt.Printf("write batch %+v - success\n", data)

	return nil
}
