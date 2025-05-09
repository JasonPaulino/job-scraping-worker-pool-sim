package main

import (
	"fmt"
	"log"
	"net/http"
)

type Job struct {
	ID  int
	URL string
}

type Result struct {
	JobID      int
	StatusCode int
	Err        error
}

func worker(id int, jobs <-chan Job, results chan<- Result) {
	_ = id

	for job := range jobs {
		resp, err := http.Get(job.URL)

		if err != nil {
			results <- Result{JobID: job.ID, StatusCode: 0, Err: err}
			continue
		}

		if resourceErr := resp.Body.Close(); resourceErr != nil {
			log.Fatal(resourceErr)
		}

		results <- Result{JobID: job.ID, StatusCode: resp.StatusCode, Err: nil}
	}
}

func main() {
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
		go worker(w, jobs, results)
	}

	// Send jobs to the job queue
	for i, url := range urls {
		jobs <- Job{ID: i + 1, URL: url}
	}

	// Collect the results
	for i := 0; i < len(urls); i++ {
		result := <-results

		if result.Err != nil {
			fmt.Printf("Job #%d FAILED: %v\n", result.JobID, result.Err)
		} else {
			fmt.Printf("Job #%d SUCCESS: status code %d\n", result.JobID, result.StatusCode)
		}
	}
}
