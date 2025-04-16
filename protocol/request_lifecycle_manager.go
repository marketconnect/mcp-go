package protocol

import (
	"context"
	"fmt"

	"sync"
	"time"
)

// TimeoutType defines the type of timeout that occurred during request processing.
type TimeoutType int

const (
	// SoftTimeout indicates that the initial soft timeout expired.
	// This is typically used to issue a cancellation notification.
	SoftTimeout TimeoutType = iota

	// MaximumTimeout indicates that the maximum allowed timeout has expired.
	// At this point, the request is forcefully cleaned up.
	MaximumTimeout
)

func (t TimeoutType) String() string {
	switch t {
	case SoftTimeout:
		return "SoftTimeout"
	case MaximumTimeout:
		return "MaximumTimeout"
	default:
		return "UnknownTimeout"
	}
}

// requestState holds the internal state of a tracked request.
// It is used internally by RequestLifecycleManager to track timeouts and activity.
type requestState[T IDConstraint] struct {
	id             IDType[T]
	softTimeout    time.Duration
	maximumTimeout time.Duration
	softTimer      *time.Timer
	maximumTimer   *time.Timer

	onTimeout    func(IDType[T], TimeoutType)
	lastActivity time.Time
}

// stop stops all active timers for the request.
func (s *requestState[T]) stop() {
	if s.softTimer != nil {
		s.softTimer.Stop()
		s.softTimer = nil
	}
	if s.maximumTimer != nil {
		s.maximumTimer.Stop()
		s.maximumTimer = nil
	}
}

// RequestLifecycleManager manages the lifecycle of MCP protocol requests.
// It enforces unique request IDs within a session and manages soft and hard timeouts.
//
// Typical usage:
//
//	manager := NewRequestLifecycleManager[string](сontext.Background())
//	err := manager.StartRequest(NewID("request-123"), 5*time.Second, 30*time.Second, func(id IDType[string], t TimeoutType) {
//	    log.Printf("Request %s timed out: %s", id, t)
//	})
//
// When a request completes successfully:
//
//	manager.CompleteRequest(NewID("request-123"))
type RequestLifecycleManager[T IDConstraint] struct {
	mu       sync.Mutex
	requests map[IDType[T]]*requestState[T]
	usedIDs  map[IDType[T]]struct{}

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup

	onError func(IDType[T], error)
}

type RequestLifecycleOption[T IDConstraint] func(*RequestLifecycleManager[T])

func WithErrorHandler[T IDConstraint](fn func(IDType[T], error)) RequestLifecycleOption[T] {
	return func(m *RequestLifecycleManager[T]) {
		m.onError = fn
	}
}

// NewRequestLifecycleManager creates and returns a new RequestLifecycleManager.
// Call StopAll() when the manager is no longer needed to clean up resources.
func NewRequestLifecycleManager[T IDConstraint](ctx context.Context, opts ...RequestLifecycleOption[T]) *RequestLifecycleManager[T] {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	manager := &RequestLifecycleManager[T]{
		requests: make(map[IDType[T]]*requestState[T]),
		usedIDs:  make(map[IDType[T]]struct{}),
		ctx:      ctx,
		cancel:   cancel,
	}

	for _, opt := range opts {
		opt(manager)
	}
	return manager
}

// Done returns a channel that's closed when the manager is stopped.
// Useful for integrating into select loops.
func (m *RequestLifecycleManager[T]) Done() <-chan struct{} {
	return m.ctx.Done()
}

// Len returns the number of currently active requests being tracked.
func (m *RequestLifecycleManager[T]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

// StartRequest begins tracking a new request with the given ID and timeout durations.
// MCP: Request IDs MUST be unique per session. This is enforced by this method.
//
// The softTimeout triggers a warning or cancellation notification if the request takes too long.
// The maximumTimeout forcefully cleans up the request state.
//
// Returns an error if:
//   - The request ID has already been used in this session.
//   - The provided timeouts are invalid.
func (m *RequestLifecycleManager[T]) StartRequest(
	id IDType[T],
	softTimeout time.Duration,
	maximumTimeout time.Duration,
	onTimeout func(IDType[T], TimeoutType),
) error {

	if onTimeout == nil {
		return ErrCallbackNil
	}

	if id.IsEmpty() {
		return ErrEmptyRequestID
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, used := m.usedIDs[id]; used {
		return ErrDuplicateRequestID
	}

	m.usedIDs[id] = struct{}{}

	if softTimeout <= 0 {

		return ErrSoftTimeoutNotPositive
	}
	if maximumTimeout <= 0 {
		return ErrMaximumTimeoutNotPositive
	}
	if softTimeout > maximumTimeout {
		return ErrSoftTimeoutExceedsMaximum
	}

	state := &requestState[T]{
		id:             id,
		softTimeout:    softTimeout,
		maximumTimeout: maximumTimeout,
		lastActivity:   time.Now(),
		onTimeout:      onTimeout,
	}

	m.wg.Add(1)

	state.softTimer = time.AfterFunc(softTimeout, func() {
		m.triggerCallback(state, SoftTimeout)
	})

	state.maximumTimer = time.AfterFunc(maximumTimeout, func() {
		m.triggerCallback(state, MaximumTimeout)
	})

	m.requests[id] = state
	return nil
}

// UpdateCallback updates the timeout callback for the specified request.
//
// Returns an error if:
//   - The request is not found.
//   - The provided callback is nil.
func (m *RequestLifecycleManager[T]) UpdateCallback(id IDType[T], newCallback func(IDType[T], TimeoutType)) error {
	if newCallback == nil {
		return ErrCallbackNil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.requests[id]
	if !exists {
		return ErrRequestNotFound
	}

	state.onTimeout = newCallback
	state.lastActivity = time.Now()
	return nil
}

// CompleteRequest stops tracking the request with the specified ID.
// Should be called when a request completes successfully.
func (m *RequestLifecycleManager[T]) CompleteRequest(id IDType[T]) {
	m.cleanupRequest(id)
}

// ResetTimeout resets the soft timeout timer for the specified request.
// Useful when receiving progress notifications to extend the active period.
func (m *RequestLifecycleManager[T]) ResetTimeout(id IDType[T]) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.requests[id]
	if !exists {
		return ErrRequestNotFound
	}

	if state.softTimer != nil {
		if !state.softTimer.Stop() {
			return nil
		}
	}

	state.softTimer = time.AfterFunc(state.softTimeout, func() {
		m.triggerCallback(state, SoftTimeout)
	})

	state.lastActivity = time.Now()
	return nil
}

// ActiveIDs returns a snapshot list of currently active request IDs.
func (m *RequestLifecycleManager[T]) ActiveIDs() []IDType[T] {
	m.mu.Lock()
	defer m.mu.Unlock()

	ids := make([]IDType[T], 0, len(m.requests))
	for id := range m.requests {
		ids = append(ids, id)
	}
	return ids
}

// StopAll stops all active requests, cancels the context, and optionally waits for all in-flight timeout callbacks to complete.
//
// Set wait=true to ensure complete deterministic shutdown before returning.
func (m *RequestLifecycleManager[T]) StopAll(wait bool) []IDType[T] {
	m.cancel()

	m.mu.Lock()
	ids := make([]IDType[T], 0, len(m.requests))
	for id, state := range m.requests {
		state.stop()
		ids = append(ids, id)
	}
	m.requests = make(map[IDType[T]]*requestState[T])
	m.mu.Unlock()

	if wait {
		m.wg.Wait()
	}

	return ids
}

// triggerCallback is an internal method that handles timeout events.
// It first checks if the manager context has been cancelled before proceeding.
func (m *RequestLifecycleManager[T]) triggerCallback(state *requestState[T], t TimeoutType) {
	select {
	case <-m.ctx.Done():
		return
	default:
	}

	m.mu.Lock()
	onTimeoutCopy := state.onTimeout
	m.mu.Unlock()

	if m.cleanupRequest(state.id) {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("callback panic: %v", r)
				if m.onError != nil {
					m.onError(state.id, err)
				} else {
					fmt.Printf("Request %v callback panicked: %v\n", state.id, r)
				}
			}
		}()

		onTimeoutCopy(state.id, t)
	}
}

// cleanupRequest stops timers and removes the request from tracking.
// Returns true if the request was found and cleaned up.
func (m *RequestLifecycleManager[T]) cleanupRequest(id IDType[T]) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.requests[id]
	if !exists {
		return false
	}

	state.stop()

	delete(m.requests, id)
	m.wg.Done()

	return true
}
