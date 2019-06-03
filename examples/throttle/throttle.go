package main

import (
	"context"
	"log"
	"sync"
	"time"

	asyncwork "github.com/GustavoKatel/asyncwork"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	worker, err := asyncwork.New()
	if err != nil {
		panic(err)
	}
	worker.Start()
	defer worker.Stop()

	worker.PostJob(func(ctx context.Context) error {
		// Long operation 1
		log.Printf("Operation1")
		wg.Done()
		return nil
	})

	worker.PostThrottledJob(func(ctx context.Context) error {
		// Long operation 2 is not executed due to throttle
		log.Printf("Operation2")
		return nil
	}, 500*time.Millisecond)

	<-time.After(600 * time.Millisecond)

	worker.PostThrottledJob(func(ctx context.Context) error {
		// Long operation 3
		log.Printf("Operation3")
		wg.Done()
		return nil
	}, 500*time.Millisecond)

	wg.Wait()
	log.Printf("Pending: %v", worker.Len())
}
