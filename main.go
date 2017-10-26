package main

import (
	"fmt"
	"time"
)

var sem = make(chan int, 10)
var count = make(chan int, 1)

func handle(r int) {
	sem <- 1  // Wait for active queue to drain.
	worker(r) // May take a long time.
	<-sem     // Done; enable next request to run.
}

func worker(i int) {
	fmt.Println("working with", i, "...")
	time.Sleep(time.Second * time.Duration(i))
	fmt.Println("done", i)
	tmp := <-count
	count <- tmp + 1
}

func main() {
	queue := make(chan int, 20)
	for i := 0; i < 20; i++ {
		queue <- i
	}
	count <- 0
	for {
		req := <-queue
		go handle(req) // Don't wait for handle to finish.

		//		if tmp := <-count; tmp == 20 {
		//			break
		//		}
	}
}
