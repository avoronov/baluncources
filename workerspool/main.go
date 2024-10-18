package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"time"
)

const (
	workersCount   = 5
	maxSleepPeriod = 1000
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	p := NewPool(workersCount)

	var (
		id      uint32
		running = true
		r       = rand.New(rand.NewPCG(1, 2))
	)

	for running {
		select {
		case <-ctx.Done():
			running = false
		default:
			id++

			j := NewJob(id, 1+r.Uint32N(maxSleepPeriod))

			if err := p.AddJob(j); err != nil {
				fmt.Printf("add job %d: %s\n", id, err)
			}
		}

		time.Sleep(time.Duration(100) * time.Millisecond)
	}

	fmt.Println("recieve stop signal - shutting down pool")

	p.Close()

	fmt.Println("pool stopped")

	p.Close() // for lulz
}
