package syncctx_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sieglu2/go_foundation/syncctx"
)

// Helper function to check if an error is a context error (timeout or cancellation)
func isContextError(err error) bool {
	return err == syncctx.ErrContextCanceled || err == syncctx.ErrContextDeadlineExceeded
}

// TestMutex tests the context-aware mutex
func TestMutex(t *testing.T) {
	t.Run("Lock with background context succeeds", func(t *testing.T) {
		var mu syncctx.Mutex
		err := mu.Lock(context.Background())
		if err != nil {
			t.Fatalf("Lock should succeed with background context, got error: %v", err)
		}
		mu.Unlock()
	})

	t.Run("Lock with canceled context fails", func(t *testing.T) {
		var mu syncctx.Mutex
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := mu.Lock(ctx)
		if !isContextError(err) {
			t.Fatalf("Lock should fail with canceled context, got: %v", err)
		}
	})

	t.Run("Lock with timeout returns error when blocked", func(t *testing.T) {
		var mu syncctx.Mutex

		// First acquire the lock
		if err := mu.Lock(context.Background()); err != nil {
			t.Fatalf("Initial lock acquisition failed: %v", err)
		}

		// Try to acquire the lock again with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := mu.Lock(ctx)
		if !isContextError(err) {
			t.Fatalf("Lock should timeout, got: %v", err)
		}

		mu.Unlock() // Unlock the initial lock
	})

	t.Run("TryLock returns false when lock is held", func(t *testing.T) {
		var mu syncctx.Mutex

		// First acquire the lock
		if err := mu.Lock(context.Background()); err != nil {
			t.Fatalf("Initial lock acquisition failed: %v", err)
		}

		// Try to acquire the lock again
		if mu.TryLock() {
			t.Fatalf("TryLock should fail when lock is held")
			mu.Unlock() // Clean up if this unexpectedly succeeds
		}

		mu.Unlock() // Unlock the initial lock
	})

	t.Run("Lock succeeds after Unlock", func(t *testing.T) {
		var mu syncctx.Mutex

		// First acquire the lock
		if err := mu.Lock(context.Background()); err != nil {
			t.Fatalf("Initial lock acquisition failed: %v", err)
		}

		// Unlock
		mu.Unlock()

		// Try to acquire the lock again - should succeed now
		if err := mu.Lock(context.Background()); err != nil {
			t.Fatalf("Lock should succeed after unlock: %v", err)
		}

		mu.Unlock() // Cleanup
	})
}

// TestRWMutex tests the context-aware RWMutex
func TestRWMutex(t *testing.T) {
	t.Run("RLock with background context succeeds", func(t *testing.T) {
		var mu syncctx.RWMutex
		err := mu.RLock(context.Background())
		if err != nil {
			t.Fatalf("RLock should succeed with background context, got error: %v", err)
		}
		mu.RUnlock()
	})

	t.Run("RLock with canceled context fails", func(t *testing.T) {
		var mu syncctx.RWMutex
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := mu.RLock(ctx)
		if !isContextError(err) {
			t.Fatalf("RLock should fail with canceled context, got: %v", err)
		}
	})

	t.Run("Lock with background context succeeds", func(t *testing.T) {
		var mu syncctx.RWMutex
		err := mu.Lock(context.Background())
		if err != nil {
			t.Fatalf("Lock should succeed with background context, got error: %v", err)
		}
		mu.Unlock()
	})

	t.Run("Lock with canceled context fails", func(t *testing.T) {
		var mu syncctx.RWMutex
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := mu.Lock(ctx)
		if !isContextError(err) {
			t.Fatalf("Lock should fail with canceled context, got: %v", err)
		}
	})

	t.Run("Lock with timeout returns error when blocked by RLock", func(t *testing.T) {
		var mu syncctx.RWMutex

		// First acquire a read lock
		if err := mu.RLock(context.Background()); err != nil {
			t.Fatalf("Initial RLock acquisition failed: %v", err)
		}

		// Try to acquire a write lock with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := mu.Lock(ctx)
		if !isContextError(err) {
			t.Fatalf("Lock should timeout when blocked by RLock, got: %v", err)
		}

		mu.RUnlock() // Unlock the initial read lock
	})

	t.Run("Multiple RLocks don't block each other", func(t *testing.T) {
		var mu syncctx.RWMutex

		// First acquire a read lock
		if err := mu.RLock(context.Background()); err != nil {
			t.Fatalf("Initial RLock acquisition failed: %v", err)
		}

		// Try to acquire another read lock - should succeed
		err := mu.RLock(context.Background())
		if err != nil {
			t.Fatalf("Second RLock should succeed, got: %v", err)
		}

		mu.RUnlock() // Unlock the second read lock
		mu.RUnlock() // Unlock the initial read lock
	})
}

// TestWaitGroup tests the context-aware wait group
func TestWaitGroup(t *testing.T) {
	t.Run("Wait with background context", func(t *testing.T) {
		var wg syncctx.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
		}()

		err := wg.Wait(context.Background())
		if err != nil {
			t.Fatalf("Wait should not fail with background context, got: %v", err)
		}
	})

	t.Run("Wait with canceled context", func(t *testing.T) {
		var wg syncctx.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(200 * time.Millisecond)
		}()

		ctx, cancel := context.WithCancel(context.Background())

		// Start waiting in a goroutine
		done := make(chan struct{})
		var waitErr error

		go func() {
			waitErr = wg.Wait(ctx)
			close(done)
		}()

		// Give the goroutine time to start waiting
		time.Sleep(50 * time.Millisecond)

		// Cancel the context
		cancel()

		select {
		case <-done:
			if !isContextError(waitErr) {
				t.Fatalf("Wait should fail with context error, got: %v", waitErr)
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("Timed out waiting for wait to return after context cancellation")
		}
	})

	t.Run("Wait with timeout that expires", func(t *testing.T) {
		var wg syncctx.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(200 * time.Millisecond)
		}()

		// Wait with a short timeout
		err := wg.WaitWithTimeout(50 * time.Millisecond)
		if !isContextError(err) {
			t.Fatalf("WaitWithTimeout should fail with timeout error, got: %v", err)
		}

		// Wait for the goroutine to finish to avoid resource leaks
		_ = wg.Wait(context.Background())
	})

	t.Run("Wait with timeout that succeeds", func(t *testing.T) {
		var wg syncctx.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
		}()

		// Wait with a longer timeout
		err := wg.WaitWithTimeout(200 * time.Millisecond)
		if err != nil {
			t.Fatalf("WaitWithTimeout should succeed, got: %v", err)
		}
	})

	t.Run("Multiple goroutines", func(t *testing.T) {
		var wg syncctx.WaitGroup

		workerCount := 5
		wg.Add(workerCount)

		for i := 0; i < workerCount; i++ {
			go func(id int) {
				defer wg.Done()
				time.Sleep(time.Duration(20+id*10) * time.Millisecond)
			}(i)
		}

		err := wg.Wait(context.Background())
		if err != nil {
			t.Fatalf("Wait should succeed with multiple goroutines, got: %v", err)
		}
	})
}

// TestRaceConditions tests for race conditions during context cancellation
func TestRaceConditions(t *testing.T) {
	t.Run("Mutex lock race", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			var mu syncctx.Mutex

			// Acquire the lock
			err := mu.Lock(context.Background())
			if err != nil {
				t.Fatalf("Initial lock failed: %v", err)
			}

			// Create a context that will be canceled
			ctx, cancel := context.WithCancel(context.Background())

			// Start a goroutine that will try to acquire the lock
			done := make(chan struct{})
			go func() {
				_ = mu.Lock(ctx)
				close(done)
			}()

			// Race between canceling the context and unlocking
			time.Sleep(time.Duration(i%10) * time.Millisecond)

			if i%2 == 0 {
				// Cancel first, then unlock
				cancel()
				time.Sleep(1 * time.Millisecond)
				mu.Unlock()
			} else {
				// Unlock first, then cancel
				mu.Unlock()
				time.Sleep(1 * time.Millisecond)
				cancel()
			}

			// Wait for the goroutine to finish
			select {
			case <-done:
				// Good, it finished
			case <-time.After(500 * time.Millisecond):
				t.Fatalf("Timed out waiting for goroutine to finish")
			}
		}
	})
}

// Benchmarks
func BenchmarkMutex(b *testing.B) {
	b.Run("Lock/Unlock", func(b *testing.B) {
		var mu syncctx.Mutex
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = mu.Lock(ctx)
			mu.Unlock()
		}
	})

	b.Run("TryLock/Unlock", func(b *testing.B) {
		var mu syncctx.Mutex

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if mu.TryLock() {
				mu.Unlock()
			}
		}
	})
}

func BenchmarkRWMutex(b *testing.B) {
	b.Run("RLock/RUnlock", func(b *testing.B) {
		var mu syncctx.RWMutex
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = mu.RLock(ctx)
			mu.RUnlock()
		}
	})

	b.Run("Lock/Unlock", func(b *testing.B) {
		var mu syncctx.RWMutex
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = mu.Lock(ctx)
			mu.Unlock()
		}
	})
}

func BenchmarkWaitGroup(b *testing.B) {
	b.Run("Add/Done/Wait", func(b *testing.B) {
		var wg syncctx.WaitGroup
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func() {
				wg.Done()
			}()
			_ = wg.Wait(ctx)
		}
	})
}

// TestParallelIncrements tests multiple goroutines incrementing a counter
func TestParallelIncrements(t *testing.T) {
	var mu syncctx.Mutex
	var counter int

	// Number of goroutines and increments per goroutine
	goroutines := 10
	incPerGoroutine := 1000

	// Total expected count
	expected := goroutines * incPerGoroutine

	// Use a wait group to wait for all goroutines
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Start goroutines
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < incPerGoroutine; j++ {
				// Lock with indefinite timeout
				if err := mu.Lock(context.Background()); err != nil {
					t.Errorf("Lock error: %v", err)
					return
				}

				// Increment counter
				counter++

				// Unlock
				mu.Unlock()
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check final counter value
	if counter != expected {
		t.Errorf("Expected counter to be %d, got %d", expected, counter)
	}
}

// TestReaderWriter tests that multiple readers can access data simultaneously
// but writers require exclusive access
func TestReaderWriter(t *testing.T) {
	var rwMu syncctx.RWMutex

	// Track active readers and writers
	var activeReaders int32
	var activeWriters int32

	// Track max concurrent readers
	var maxReaders int32

	// Number of operations
	readers := 100
	writers := 10

	// Use a wait group to wait for all operations
	var wg sync.WaitGroup
	wg.Add(readers + writers)

	// Start reader goroutines
	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()

			// Acquire read lock
			if err := rwMu.RLock(context.Background()); err != nil {
				t.Errorf("RLock error: %v", err)
				return
			}

			// Increment active readers
			readers := atomic.AddInt32(&activeReaders, 1)

			// Update max readers
			for {
				current := atomic.LoadInt32(&maxReaders)
				if readers <= current || atomic.CompareAndSwapInt32(&maxReaders, current, readers) {
					break
				}
			}

			// Check that no writers are active
			if atomic.LoadInt32(&activeWriters) > 0 {
				t.Errorf("Writer active during read")
			}

			// Simulate read operation
			time.Sleep(time.Millisecond)

			// Decrement active readers
			atomic.AddInt32(&activeReaders, -1)

			// Release read lock
			rwMu.RUnlock()
		}()
	}

	// Start writer goroutines
	for i := 0; i < writers; i++ {
		go func() {
			defer wg.Done()

			// Acquire write lock
			if err := rwMu.Lock(context.Background()); err != nil {
				t.Errorf("Lock error: %v", err)
				return
			}

			// Increment active writers
			writers := atomic.AddInt32(&activeWriters, 1)

			// Check that only one writer is active
			if writers != 1 {
				t.Errorf("Multiple writers active")
			}

			// Check that no readers are active
			if atomic.LoadInt32(&activeReaders) > 0 {
				t.Errorf("Readers active during write")
			}

			// Simulate write operation
			time.Sleep(time.Millisecond)

			// Decrement active writers
			atomic.AddInt32(&activeWriters, -1)

			// Release write lock
			rwMu.Unlock()
		}()
	}

	// Wait for all operations to finish
	wg.Wait()

	// Check that multiple readers were active concurrently
	if maxReaders <= 1 {
		t.Errorf("Expected multiple concurrent readers, got %d", maxReaders)
	}
}

// TestWaitGroupCancellation tests that a WaitGroup's Wait can be canceled
func TestWaitGroupCancellation(t *testing.T) {
	var wg syncctx.WaitGroup

	// Add a task that will never complete
	wg.Add(1)

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Start waiting in a goroutine
	start := time.Now()
	err := wg.Wait(ctx)
	elapsed := time.Since(start)

	// Check that wait returned with a context error
	if !isContextError(err) {
		t.Errorf("Expected context error, got %v", err)
	}

	// Check that wait returned within a reasonable time
	if elapsed < 50*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Errorf("Wait returned in unexpected time: %v", elapsed)
	}
}
