package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JasonPaulino/job-scraping-worker-pool-sim/internal/job"
	"github.com/google/uuid"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	defer cancel()

	var wg sync.WaitGroup

	urls := []string{
		"https://google.com",
		"https://github.com",
		"https://nonexistent.website.abc", // simulate a failure
		"https://golang.org",
	}

	jobs := make(chan job.Job, len(urls))
	results := make(chan job.Result, len(urls))

	numOfWorkers := 3

	// Start workers
	for w := 1; w <= numOfWorkers; w++ {
		workerUUID := uuid.New().String()
		wg.Add(1)
		go job.Worker(ctx, workerUUID, jobs, results, &wg)
	}

	// Send jobs to the job queue
	for _, url := range urls {
		jobs <- job.Job{ID: uuid.New().String(), URL: url}
	}

	close(jobs)
	wg.Wait()
	close(results)

	// Collect the results
	for result := range results {
		if result.Errors != nil {
			fmt.Printf("Job %s FAILED: %v\n", result.JobID, result.Errors)
		} else {
			fmt.Printf("Job %s SUCCESS: status code %d\n", result.JobID, result.StatusCode)
		}
	}
}
