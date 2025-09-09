package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8080", "API endpoint to test")
	requests := flag.Int("n", 100, "Number of requests to send")
	concurrency := flag.Int("c", 10, "Number of concurrent workers")
	flag.Parse()

	   var wg sync.WaitGroup
	   jobs := make(chan int, *requests)
	   results := make(chan time.Duration, *requests)
	   errors := make(chan error, *requests)

	   start := time.Now()

	   worker := func(id int) {
			   defer wg.Done()
			   for job := range jobs {
					   fmt.Printf("Worker %d sending request #%d to %s\n", id, job+1, *url)
					   begin := time.Now()
					   resp, err := http.Get(*url)
					   if err != nil {
							   errors <- err
							   continue
					   }
					   resp.Body.Close()
					   results <- time.Since(begin)
			   }
	   }

	   for i := 0; i < *concurrency; i++ {
			   wg.Add(1)
			   go worker(i)
	   }

	   for i := 0; i < *requests; i++ {
			   jobs <- i
	   }
	   close(jobs)

	   wg.Wait()
	   close(results)
	   close(errors)
	   totalTime := time.Since(start)

	   success := 0
	   failures := 0
	   var sum time.Duration
	   for d := range results {
			   sum += d
			   success++
	   }
	   for range errors {
			   failures++
	   }
	   avg := time.Duration(0)
	   if success > 0 {
			   avg = sum / time.Duration(success)
	   }

	   fmt.Printf("Total requests: %d\n", *requests)
	   fmt.Printf("Success: %d\n", success)
	   fmt.Printf("Failures: %d\n", failures)
	   fmt.Printf("Total time: %v\n", totalTime)
	   fmt.Printf("Average response time: %v\n", avg)
}
