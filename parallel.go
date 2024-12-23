package foundation

import (
	"fmt"
	"sync"
	"time"
)

func RunInParallel(numWorkers int, sleepIntermittent time.Duration, params chan any, op func(any) error) error {
	if numWorkers <= 0 {
		return fmt.Errorf("number of workers must be positive")
	}

	// Create a WaitGroup to track when all workers are done
	var wg sync.WaitGroup

	// Create error channel to collect errors from workers
	errChan := make(chan error)

	// Start workers
	for i := 0; i < numWorkers; i += 1 {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for param := range params {
				defer time.Sleep(sleepIntermittent)

				if err := op(param); err != nil {
					errChan <- fmt.Errorf("worker %d failed processing %+v: %v", workerID, param, err)
					return
				}
			}
		}(i)
	}

	// Wait for all workers to finish
	wg.Wait()
	close(errChan)

	// Collect any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// If there were any errors, return them combined
	if len(errors) > 0 {
		return fmt.Errorf("processing errors: %v", errors)
	}

	return nil
}
