package protocol_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/marketconnect/mcp-go/protocol"
)

type testID = string

func newTestID(s string) protocol.IDType[testID] {
	return protocol.NewID[testID](s)
}

func waitForGeneric[T any](ch <-chan T, timeout time.Duration) (T, bool) {
	var zero T
	select {
	case v := <-ch:
		return v, true
	case <-time.After(timeout):
		return zero, false
	}
}

func TestRequestLifecycleSuccessStartAndComplete(t *testing.T) {
	var timeoutCalled bool
	id := newTestID("req-1")

	mgr := protocol.NewRequestLifecycleManager[testID](context.Background())

	err := mgr.StartRequest(id, 50*time.Millisecond, 200*time.Millisecond, func(_ protocol.IDType[testID], typ protocol.TimeoutType) {
		timeoutCalled = true
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mgr.Len() != 1 {
		t.Errorf("expected 1 active request, got %d", mgr.Len())
	}

	mgr.CompleteRequest(id)
	time.Sleep(100 * time.Millisecond)

	if timeoutCalled {
		t.Errorf("timeout callback should not be called after complete")
	}
}

func TestRequestLifecycleDuplicateID(t *testing.T) {
	id := newTestID("dup")
	mgr := protocol.NewRequestLifecycleManager[testID](context.Background())

	if err := mgr.StartRequest(id, time.Second, 2*time.Second, func(_ protocol.IDType[testID], _ protocol.TimeoutType) {

	}); err != nil {
		t.Fatalf("unexpected error on first StartRequest: %v", err)
	}

	err := mgr.StartRequest(id, time.Second, 2*time.Second, func(_ protocol.IDType[testID], _ protocol.TimeoutType) {

	})
	if !errors.Is(err, protocol.ErrDuplicateRequestID) {
		t.Errorf("expected ErrDuplicateRequestID, got: %v", err)
	}
}

func TestRequestLifecycleSoftTimeoutFires(t *testing.T) {
	id := newTestID("timeout-soft")
	fired := make(chan protocol.TimeoutType, 1)

	mgr := protocol.NewRequestLifecycleManager[testID](context.Background())
	err := mgr.StartRequest(id, 30*time.Millisecond, 200*time.Millisecond, func(_ protocol.IDType[testID], typ protocol.TimeoutType) {
		fired <- typ
	})
	if err != nil {
		t.Fatalf("start error: %v", err)
	}

	typ, ok := waitForGeneric(fired, 100*time.Millisecond)
	if !ok {
		t.Fatalf("soft timeout did not fire")
	}
	if typ != protocol.SoftTimeout {
		t.Errorf("expected SoftTimeout, got %v", typ)
	}
}

func TestRequestLifecycleMaxTimeoutFires(t *testing.T) {
	id := newTestID("timeout-max")
	fired := make(chan protocol.TimeoutType, 1)

	mgr := protocol.NewRequestLifecycleManager[testID](context.Background())
	err := mgr.StartRequest(id, 30*time.Millisecond, 30*time.Millisecond, func(_ protocol.IDType[testID], typ protocol.TimeoutType) {
		fired <- typ
	})
	if err != nil {
		t.Fatalf("start error: %v", err)
	}

	typ, ok := waitForGeneric(fired, 100*time.Millisecond)
	if !ok {
		t.Fatalf("timeout did not fire")
	}
	if typ != protocol.MaximumTimeout && typ != protocol.SoftTimeout {
		t.Errorf("expected a timeout (Soft or Max), got: %v", typ)
	}
}

func TestRequestLifecycleUpdateCallback(t *testing.T) {
	id := newTestID("change-cb")

	var mu sync.Mutex
	events := []string{}

	cb1 := func(_ protocol.IDType[testID], _ protocol.TimeoutType) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, "first")
	}

	cb2 := func(_ protocol.IDType[testID], _ protocol.TimeoutType) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, "second")
	}

	mgr := protocol.NewRequestLifecycleManager[testID](context.Background())
	_ = mgr.StartRequest(id, 100*time.Millisecond, 200*time.Millisecond, cb1)

	time.Sleep(50 * time.Millisecond)
	_ = mgr.UpdateCallback(id, cb2)
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 1 || events[0] != "second" {
		t.Errorf("expected only second callback to fire, got: %v", events)
	}
}

func TestRequestLifecycleResetTimeout(t *testing.T) {
	id := newTestID("reset-test")
	fired := make(chan protocol.TimeoutType, 1)

	mgr := protocol.NewRequestLifecycleManager[testID](context.Background())
	_ = mgr.StartRequest(id, 40*time.Millisecond, 200*time.Millisecond, func(_ protocol.IDType[testID], typ protocol.TimeoutType) {
		fired <- typ
	})

	time.Sleep(30 * time.Millisecond)
	_ = mgr.ResetTimeout(id)

	if _, ok := waitForGeneric(fired, 30*time.Millisecond); ok {
		t.Fatalf("timeout fired too early after reset")
	}
	if _, ok := waitForGeneric(fired, 100*time.Millisecond); !ok {
		t.Fatalf("timeout never fired after reset")
	}
}

func TestRequestLifecycleStopAll(t *testing.T) {
	id := newTestID("stopall")
	fired := make(chan protocol.TimeoutType, 1)

	mgr := protocol.NewRequestLifecycleManager[testID](context.Background())
	_ = mgr.StartRequest(id, 50*time.Millisecond, 100*time.Millisecond, func(_ protocol.IDType[testID], typ protocol.TimeoutType) {
		fired <- typ
	})

	mgr.StopAll(false)

	if _, ok := waitForGeneric(fired, 100*time.Millisecond); ok {
		t.Errorf("callback should not fire after StopAll")
	}
}

func TestRequestLifecycleCallbackPanicRecovery(t *testing.T) {
	id := newTestID("panic")
	panicked := make(chan struct{})

	mgr := protocol.NewRequestLifecycleManager[testID](context.Background(), protocol.WithErrorHandler(func(id protocol.IDType[testID], err error) {
		panicked <- struct{}{}
	}))

	_ = mgr.StartRequest(id, 10*time.Millisecond, 50*time.Millisecond, func(_ protocol.IDType[testID], _ protocol.TimeoutType) {
		panic("boom")
	})

	if _, ok := waitForGeneric(panicked, 100*time.Millisecond); !ok {
		t.Fatalf("panic recovery handler was not called")
	}
}

func TestTimeoutTypeString(t *testing.T) {
	tests := []struct {
		tt   protocol.TimeoutType
		want string
	}{
		{protocol.SoftTimeout, "SoftTimeout"},
		{protocol.MaximumTimeout, "MaximumTimeout"},
		{protocol.TimeoutType(999), "UnknownTimeout"},
	}

	for _, tc := range tests {
		got := tc.tt.String()
		if got != tc.want {
			t.Errorf("TimeoutType(%d).String() = %q; want %q", tc.tt, got, tc.want)
		}
	}
}

func TestRequestLifecycleManagerDone(t *testing.T) {
	manager := protocol.NewRequestLifecycleManager[string](context.Background())
	doneCh := manager.Done()

	select {
	case <-doneCh:
		t.Fatal("Expected Done channel to be open initially")
	default:
	}

	manager.StopAll(false)

	select {
	case <-doneCh:

	default:
		t.Fatal("Expected Done channel to be closed after StopAll")
	}
}

func TestStartRequestErrors(t *testing.T) {
	manager := protocol.NewRequestLifecycleManager[string](context.Background())

	tests := []struct {
		name           string
		id             protocol.IDType[string]
		softTimeout    time.Duration
		maximumTimeout time.Duration
		onTimeout      func(protocol.IDType[string], protocol.TimeoutType)
		wantErr        error
	}{
		{
			name:           "nil callback",
			id:             protocol.NewID("req-nil"),
			softTimeout:    time.Second,
			maximumTimeout: 2 * time.Second,
			onTimeout:      nil,
			wantErr:        protocol.ErrCallbackNil,
		},
		{
			name:           "empty ID",
			id:             protocol.IDType[string]{},
			softTimeout:    time.Second,
			maximumTimeout: 2 * time.Second,
			onTimeout:      func(protocol.IDType[string], protocol.TimeoutType) {},
			wantErr:        protocol.ErrEmptyRequestID,
		},
		{
			name:           "non-positive soft timeout",
			id:             protocol.NewID("req-soft0"),
			softTimeout:    0,
			maximumTimeout: time.Second,
			onTimeout:      func(protocol.IDType[string], protocol.TimeoutType) {},
			wantErr:        protocol.ErrSoftTimeoutNotPositive,
		},
		{
			name:           "non-positive maximum timeout",
			id:             protocol.NewID("req-max0"),
			softTimeout:    time.Second,
			maximumTimeout: 0,
			onTimeout:      func(protocol.IDType[string], protocol.TimeoutType) {},
			wantErr:        protocol.ErrMaximumTimeoutNotPositive,
		},
		{
			name:           "soft timeout exceeds maximum",
			id:             protocol.NewID("req-bad"),
			softTimeout:    3 * time.Second,
			maximumTimeout: 2 * time.Second,
			onTimeout:      func(protocol.IDType[string], protocol.TimeoutType) {},
			wantErr:        protocol.ErrSoftTimeoutExceedsMaximum,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.StartRequest(tc.id, tc.softTimeout, tc.maximumTimeout, tc.onTimeout)
			if err != tc.wantErr {
				t.Errorf("Expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestUpdateCallbackErrors(t *testing.T) {
	manager := protocol.NewRequestLifecycleManager[string](context.Background())
	id := protocol.NewID("req-1")

	err := manager.UpdateCallback(id, func(protocol.IDType[string], protocol.TimeoutType) {})
	if err != protocol.ErrRequestNotFound {
		t.Errorf("Expected ErrRequestNotFound, got %v", err)
	}

	err = manager.StartRequest(id, time.Second, 2*time.Second, func(protocol.IDType[string], protocol.TimeoutType) {})
	if err != nil {
		t.Fatalf("Failed to start request: %v", err)
	}

	err = manager.UpdateCallback(id, nil)
	if err != protocol.ErrCallbackNil {
		t.Errorf("Expected ErrCallbackNil, got %v", err)
	}
}

func TestResetTimeoutErrors(t *testing.T) {
	manager := protocol.NewRequestLifecycleManager[string](context.Background())

	id := protocol.NewID("not-exist")
	err := manager.ResetTimeout(id)
	if err != protocol.ErrRequestNotFound {
		t.Errorf("Expected ErrRequestNotFound, got %v", err)
	}
}

func TestResetTimeoutStopReturnsFalse(t *testing.T) {
	manager := protocol.NewRequestLifecycleManager[string](context.Background())

	id := protocol.NewID("req1")

	blocker := make(chan struct{})

	err := manager.StartRequest(id, 200*time.Millisecond, 2*time.Second, func(protocol.IDType[string], protocol.TimeoutType) {
		<-blocker
	})
	if err != nil {
		t.Fatalf("start request failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	err = manager.ResetTimeout(id)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	close(blocker)
}

func TestActiveIDs(t *testing.T) {
	manager := protocol.NewRequestLifecycleManager[string](context.Background())

	id1 := protocol.NewID("id-1")
	id2 := protocol.NewID("id-2")

	manager.StartRequest(id1, time.Second, 2*time.Second, func(protocol.IDType[string], protocol.TimeoutType) {})
	manager.StartRequest(id2, time.Second, 2*time.Second, func(protocol.IDType[string], protocol.TimeoutType) {})

	ids := manager.ActiveIDs()
	if len(ids) != 2 {
		t.Fatalf("Expected 2 active IDs, got %d", len(ids))
	}

	found := map[string]bool{}
	for _, id := range ids {
		found[id.String()] = true
	}

	if !found["id-1"] || !found["id-2"] {
		t.Errorf("Expected both IDs in active list, got %v", ids)
	}
}

func TestResetTimeoutStopReturnsFalsePath(t *testing.T) {
	manager := protocol.NewRequestLifecycleManager[string](context.Background())

	id := protocol.NewID("soft-stop-false")
	block := make(chan struct{})

	err := manager.StartRequest(id, 150*time.Millisecond, 2*time.Second, func(protocol.IDType[string], protocol.TimeoutType) {
		<-block
	})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	err = manager.ResetTimeout(id)
	if err != nil {
		t.Errorf("Expected no error when Stop() returns false, got %v", err)
	}

	close(block)
}
