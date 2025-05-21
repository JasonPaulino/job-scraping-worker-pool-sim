package job

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func Worker(id string, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	_ = id // id param isn't used so we satisfy compiler

	defer wg.Done()

	var (
		finalResp *http.Response
		errors    []error
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
			errors = append(errors, err) // Resetting to latest error

			fmt.Printf("[Worker %s] Attempt %d failed for job %s: %v\n", id, attempt, job.ID, err)
			time.Sleep(500 * time.Millisecond) // backoff before retry
		}

		if finalResp != nil {
			defer finalResp.Body.Close()

			results <- Result{
				JobID:      job.ID,
				StatusCode: finalResp.StatusCode,
				Errors:     nil,
			}
		} else {
			results <- Result{
				JobID:      job.ID,
				StatusCode: 0,
				Errors:     errors,
			}
		}
	}
}
