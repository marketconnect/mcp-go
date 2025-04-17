package protocol

import (
	"context"
	"testing"
	"time"
)

func TestNewRequestLifecycleManagerNilContext(t *testing.T) {
	manager := NewRequestLifecycleManager[string](nil)
	if manager.ctx == nil {
		t.Fatal("Expected non-nil context")
	}
	select {
	case <-manager.ctx.Done():
		t.Fatal("Expected context to be active, but it is done")
	default:
	}
}
func TestTriggerCallbackContextAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	manager := &RequestLifecycleManager[string]{
		ctx:    ctx,
		cancel: func() {},
	}

	state := &requestState[string]{
		id:        NewID("ctx-done"),
		onTimeout: func(IDType[string], TimeoutType) {},
	}

	manager.triggerCallback(state, SoftTimeout)

}

func TestTriggerCallbackPanicNoErrorHandler(t *testing.T) {
	manager := NewRequestLifecycleManager[string](context.Background())

	id := NewID("panic-no-handler")

	state := &requestState[string]{
		id:        id,
		onTimeout: func(IDType[string], TimeoutType) { panic("boom!") },
	}

	manager.requests[id] = state
	manager.wg.Add(1)

	manager.triggerCallback(state, SoftTimeout)

	manager.wg.Wait()
}
func TestStopAllWaitTriggersWaitGroup(t *testing.T) {
	ctx := context.Background()
	manager := NewRequestLifecycleManager[string](ctx)

	id := NewID("test-stopall-wait")

	triggered := make(chan struct{})
	blockDone := make(chan struct{})

	state := &requestState[string]{
		id: id,
		onTimeout: func(IDType[string], TimeoutType) {
			close(triggered)
			<-blockDone
		},
	}

	manager.requests[id] = state
	manager.wg.Add(1)

	go func() {
		manager.triggerCallback(state, SoftTimeout)
	}()

	<-triggered

	start := time.Now()
	done := make(chan struct{})
	go func() {
		manager.StopAll(true)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	close(blockDone)

	<-done
	elapsed := time.Since(start)

	if elapsed < 100*time.Millisecond {
		t.Errorf("StopAll returned too early: %v", elapsed)
	}
}
func TestResetTimeout_TimerStopReturnsFalse(t *testing.T) {
	manager := NewRequestLifecycleManager[string](context.Background())

	id := NewID("reset-false-stop")

	timer := time.NewTimer(1 * time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	state := &requestState[string]{
		id:        id,
		softTimer: timer,
		onTimeout: func(IDType[string], TimeoutType) {},
	}

	manager.requests[id] = state
	manager.usedIDs[id] = struct{}{}
	manager.wg.Add(1)

	err := manager.ResetTimeout(id)
	if err != nil {
		t.Errorf("Expected no error when Stop returns false, got: %v", err)
	}

	manager.cleanupRequest(id)
}
