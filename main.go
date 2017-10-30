package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var count int32

func generateData(ctx context.Context, queue chan<- int, maxData int) {
	for i := 0; i < maxData; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("generateData stopping...")
			close(queue)
			return
		default:
			queue <- i
		}
	}
	close(queue)
}

func worker(ctx context.Context, i int, queue <-chan int) {
	for value := range queue {
		select {
		case <-ctx.Done():
			fmt.Println("worker ", i, "stopping...")
			return
		default:
			fmt.Println("worker", i, "working with", value, "...")
			time.Sleep(1 * time.Second)
			fmt.Println("done", value)
			atomic.AddInt32(&count, 1)
		}
	}
}

type workerFunc func(ctx context.Context, id int, queue <-chan int)

func runWorkers(ctx context.Context, queue <-chan int, numWorkers int, worker workerFunc) {
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			worker(ctx, i, queue)
		}(i)
	}
	wg.Wait()
}

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		var sigCount int
		for sig := range sigChan {
			sigCount++
			switch sigCount {
			case 1:
				fmt.Printf("First signal %+v received. Cancelling...\n", sig)
				cancelFunc()
			default:
				fmt.Printf("Another signal %+v received. Stopping...\n", sig)
				os.Exit(1)
			}
		}
	}()

	queue := make(chan int, 20)
	go generateData(ctx, queue, 100)

	runWorkers(ctx, queue, 10, worker)

	fmt.Println(count)
}
