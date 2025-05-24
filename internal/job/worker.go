package job

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func Worker(ctx context.Context, id string, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	_ = id // id param isn't used so we satisfy compiler

	defer wg.Done()

	var (
		finalResp *http.Response
		errors    []error
	)

	for job := range jobs {
		maxRetries := 3

		for attempt := 1; attempt <= maxRetries; attempt++ {
			select {
			case <-ctx.Done(): // blocks until timeout or cancel
				// Context canceled or timeout exceeded
				errors = append(errors, ctx.Err())
				fmt.Printf("[Worker %s] Job %s canceled due to: %v\n", id, job.ID, ctx.Err())
				break // continues to next job. "return" if we want the worker to stop working fully
			default:
				// Normal retry attempt
				req, _ := http.NewRequestWithContext(ctx, "GET", job.URL, nil)
				resp, err := http.DefaultClient.Do(req)

				if err == nil { // Response was successful
					finalResp = resp
					break
				}

				errors = append(errors, err)

				fmt.Printf("[Worker %s] Attempt %d failed for job %s: %v\n", id, attempt, job.ID, err)
				time.Sleep(500 * time.Millisecond) // backoff before retry
			}
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
