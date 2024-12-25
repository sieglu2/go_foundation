package foundation

import (
	"fmt"
	"sync"
	"time"
)

func RunInParallel(
	numWorkers int, sleepIntermittent time.Duration, params chan any,
	onRun func(any) error, onComplete func([]error) error,
) error {
	if numWorkers <= 0 {
		return fmt.Errorf("number of workers must be positive")
	}

	var wg sync.WaitGroup

	errChan := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i += 1 {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for param := range params {
				if err := onRun(param); err != nil {
					errChan <- fmt.Errorf("worker %d failed processing %+v: %v", workerID, param, err)
				}

				time.Sleep(sleepIntermittent)
			}
		}(i)
	}

	// Start a goroutine to close errChan after all workers finish
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// Run onComplete after collecting errors
	if err := onComplete(errors); err != nil {
		return fmt.Errorf("failed on onComplete: %v", err)
	}

	// If there were any errors, return them combined
	if len(errors) > 0 {
		return fmt.Errorf("processing errors: %v", errors)
	}

	return nil
}
