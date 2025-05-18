package main

import (
	"github.com/google/uuid"
	"sync"
	"time"

	"fmt"
	"net/http"
)

type Job struct {
	ID  string
	URL string
}

type Result struct {
	JobID      string
	StatusCode int
	Err        error
}

func worker(id string, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	_ = id // id param isn't used so we satisfy compiler

	defer wg.Done()

	var (
		finalResp   *http.Response
		finalJobErr error
	)

	for job := range jobs {
		maxRetries := 3

		for attempt := 1; attempt <= maxRetries; attempt++ {
			resp, err := http.Get(job.URL)

			if err == nil { // Response was successful
				finalResp = resp
				break
			}

			/*
				NOTE: We could change the Result struct Err field
				To a list of error so we capture every retry
			*/
			finalJobErr = err // Resetting to latest error

			fmt.Printf("[Worker %s] Attempt %d failed for job %s: %v\n", id, attempt, job.ID, err)
			time.Sleep(500 * time.Millisecond) // backoff before retry
		}

		if finalResp != nil {
			defer finalResp.Body.Close()

			results <- Result{
				JobID:      job.ID,
				StatusCode: finalResp.StatusCode,
				Err:        nil,
			}
		} else {
			results <- Result{
				JobID:      job.ID,
				StatusCode: 0,
				Err:        finalJobErr,
			}
		}
	}
}

func main() {
	var wg sync.WaitGroup

	urls := []string{
		"https://google.com",
		"https://github.com",
		"https://nonexistent.website.abc", // simulate a failure
		"https://golang.org",
	}

	jobs := make(chan Job, len(urls))
	results := make(chan Result, len(urls))

	numOfWorkers := 3

	// Start workers
	for w := 1; w <= numOfWorkers; w++ {
		workerUUID := uuid.New().String()
		wg.Add(1)
		go worker(workerUUID, jobs, results, &wg)
	}

	// Send jobs to the job queue
	for _, url := range urls {
		jobUUID := uuid.New().String()

		jobs <- Job{ID: jobUUID, URL: url}
	}

	close(jobs)
	wg.Wait()
	close(results)

	// Collect the results
	for i := 0; i < len(urls); i++ {
		result := <-results

		if result.Err != nil {
			fmt.Printf("Job %s FAILED: %v\n", result.JobID, result.Err)
		} else {
			fmt.Printf("Job %s SUCCESS: status code %d\n", result.JobID, result.StatusCode)
		}
	}
}
