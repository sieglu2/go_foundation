package foundation

import (
	"fmt"
	"sync"
	"time"
)

func RunInParallel(
	numWorkers int, sleepIntermittent time.Duration,
	params chan any,
	onRun func(any) error, onComplete func([]error) error,
) error {
	if numWorkers <= 0 {
		return fmt.Errorf("number of workers must be positive")
	}

	var wg sync.WaitGroup
	errChan := make(chan error) // Unbuffered error channel

	// Start a goroutine to collect errors
	errorCollectorDone := make(chan struct{})
	var errors []error
	go func() {
		defer close(errorCollectorDone)
		for err := range errChan {
			errors = append(errors, err)
		}
	}()

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
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

	// Wait for all workers to finish, then close error channel
	wg.Wait()
	close(errChan)

	// Wait for error collector to finish
	<-errorCollectorDone

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
