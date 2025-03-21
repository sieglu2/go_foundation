package syncctx

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrContextCanceled is returned when the context is canceled
	ErrContextCanceled = errors.New("context canceled")
	// ErrContextDeadlineExceeded is returned when the context deadline is exceeded
	ErrContextDeadlineExceeded = errors.New("context deadline exceeded")
)

// ContextError converts a context error to a specific error
func ContextError(ctx context.Context) error {
	if ctx.Err() == context.Canceled {
		return ErrContextCanceled
	}
	if ctx.Err() == context.DeadlineExceeded {
		return ErrContextDeadlineExceeded
	}
	return ctx.Err()
}

// Mutex is a context-aware mutex
type Mutex struct {
	mu sync.Mutex
}

// Lock acquires the lock, respecting context cancellation
func (m *Mutex) Lock(ctx context.Context) error {
	// Create a channel to detect lock acquisition
	acquired := make(chan struct{})

	// Try to acquire the lock in a separate goroutine
	go func() {
		m.mu.Lock()
		close(acquired)
	}()

	// Wait for either lock acquisition or context cancellation
	select {
	case <-acquired:
		return nil
	case <-ctx.Done():
		// Note: we can't cancel the lock acquisition, but we can
		// return immediately to indicate cancellation
		return ContextError(ctx)
	}
}

// TryLock attempts to acquire the lock without blocking
func (m *Mutex) TryLock() bool {
	// Use a minimal timeout to avoid blocking
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	// Try to lock with a tiny timeout
	return m.Lock(ctx) == nil
}

// LockWithTimeout attempts to acquire the lock with a timeout
func (m *Mutex) LockWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return m.Lock(ctx)
}

// Unlock releases the lock
func (m *Mutex) Unlock() {
	m.mu.Unlock()
}

// RWMutex is a context-aware read-write mutex
type RWMutex struct {
	mu sync.RWMutex
}

// Lock acquires a write lock, respecting context cancellation
func (rw *RWMutex) Lock(ctx context.Context) error {
	// Create a channel to detect lock acquisition
	acquired := make(chan struct{})

	// Try to acquire the lock in a separate goroutine
	go func() {
		rw.mu.Lock()
		close(acquired)
	}()

	// Wait for either lock acquisition or context cancellation
	select {
	case <-acquired:
		return nil
	case <-ctx.Done():
		// Note: we can't cancel the lock acquisition, but we can
		// return immediately to indicate cancellation
		return ContextError(ctx)
	}
}

// TryLock attempts to acquire a write lock without blocking
func (rw *RWMutex) TryLock() bool {
	// Use a minimal timeout to avoid blocking
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	// Try to lock with a tiny timeout
	return rw.Lock(ctx) == nil
}

// LockWithTimeout attempts to acquire a write lock with a timeout
func (rw *RWMutex) LockWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return rw.Lock(ctx)
}

// Unlock releases a write lock
func (rw *RWMutex) Unlock() {
	rw.mu.Unlock()
}

// RLock acquires a read lock, respecting context cancellation
func (rw *RWMutex) RLock(ctx context.Context) error {
	// Create a channel to detect lock acquisition
	acquired := make(chan struct{})

	// Try to acquire the lock in a separate goroutine
	go func() {
		rw.mu.RLock()
		close(acquired)
	}()

	// Wait for either lock acquisition or context cancellation
	select {
	case <-acquired:
		return nil
	case <-ctx.Done():
		// Note: we can't cancel the lock acquisition, but we can
		// return immediately to indicate cancellation
		return ContextError(ctx)
	}
}

// TryRLock attempts to acquire a read lock without blocking
func (rw *RWMutex) TryRLock() bool {
	// Use a minimal timeout to avoid blocking
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	// Try to lock with a tiny timeout
	return rw.RLock(ctx) == nil
}

// RLockWithTimeout attempts to acquire a read lock with a timeout
func (rw *RWMutex) RLockWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return rw.RLock(ctx)
}

// RUnlock releases a read lock
func (rw *RWMutex) RUnlock() {
	rw.mu.RUnlock()
}

// WaitGroup is a context-aware wait group
type WaitGroup struct {
	wg sync.WaitGroup
}

// Add adds delta to the WaitGroup counter
func (wg *WaitGroup) Add(delta int) {
	wg.wg.Add(delta)
}

// Done decrements the WaitGroup counter
func (wg *WaitGroup) Done() {
	wg.wg.Done()
}

// Wait waits for the WaitGroup counter to be zero, respecting context cancellation
func (wg *WaitGroup) Wait(ctx context.Context) error {
	// Create a channel to detect completion
	done := make(chan struct{})

	// Wait in a separate goroutine
	go func() {
		wg.wg.Wait()
		close(done)
	}()

	// Wait for either completion or context cancellation
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ContextError(ctx)
	}
}

// WaitWithTimeout waits for the WaitGroup counter to be zero with a timeout
func (wg *WaitGroup) WaitWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return wg.Wait(ctx)
}
