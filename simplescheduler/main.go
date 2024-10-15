package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func main() {
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	// defer cancel()

	sched := NewSimpleScheduler()

	start := time.Now()

	sched.SetTimeout(1*time.Second, func() { fmt.Printf("first job done (%f sec elapsed)\n", time.Since(start).Seconds()) }) // firstJobID :=
	secondJobID := sched.SetTimeout(2*time.Second, func() { fmt.Printf("second job done (%f sec elapsed)\n", time.Since(start).Seconds()) })
	sched.SetTimeout(3*time.Second, func() { fmt.Printf("third job done (%f sec elapsed)\n", time.Since(start).Seconds()) }) // thirdJobID :=
	fourthJobID := sched.SetTimeout(4*time.Second, func() { fmt.Printf("fourth job done (%f sec elapsed)\n", time.Since(start).Seconds()) })
	sched.SetTimeout(5*time.Second, func() { fmt.Printf("fifth job done (%f sec elapsed)\n", time.Since(start).Seconds()) }) // fifthJobID :=

	for _, id := range []uuid.UUID{secondJobID, fourthJobID} {
		if err := sched.CancelTimeout(id); err != nil {
			fmt.Printf("cancel job %s: %v\n", id, err)
		} else {
			fmt.Printf("job %s cancelled\n", id)
		}
	}

	for _, id := range []uuid.UUID{secondJobID, fourthJobID} {
		if err := sched.CancelTimeout(id); err != nil {
			fmt.Printf("try to cancel job %s twice: %v\n", id, err)
		}
	}

	time.Sleep(7 * time.Second)

	fmt.Printf("all non-cancelled jobs should be done at that moment\n")
}
