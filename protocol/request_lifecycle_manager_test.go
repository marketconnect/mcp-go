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

	_ = mgr.StartRequest(id, time.Second, 2*time.Second, func(_ protocol.IDType[testID], _ protocol.TimeoutType) {})

	err := mgr.StartRequest(id, time.Second, 2*time.Second, func(_ protocol.IDType[testID], _ protocol.TimeoutType) {})
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

	// Ожидаем максимум таймаут
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
