//go:build !coverage
// +build !coverage

package protocol

import (
	"testing"
	"time"
)

// Эти тесты могут вызывать зависания при проверке покрытия,
// поэтому они исключены из стандартного запуска с coverage.
// Для их запуска используйте: go test -tags=flaky ./protocol/...

// TestStopAllWaitsForCallbacks проверяет, что StopAll ожидает завершения всех колбэков
func TestStopAllWaitsForCallbacks_Flaky(t *testing.T) {
	manager := NewRequestLifecycleManager[string](nil)
	id := newID("stop-all-sync")

	started := make(chan struct{})
	finished := make(chan struct{})

	callback := func(ID[string], TimeoutType) {
		close(started)
		time.Sleep(100 * time.Millisecond)
		close(finished)
	}

	manager.StartRequest(id, time.Hour, 2*time.Hour, callback)

	state := manager.requests[id]

	doneTrigger := make(chan struct{})

	go func() {
		manager.triggerCallback(state, SoftTimeout)
		close(doneTrigger)
	}()

	select {
	case <-started:
		// OK, колбэк начался
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Callback did not start in time")
	}

	start := time.Now()

	go func() {
		<-doneTrigger
		manager.StopAll(true)
	}()

	select {
	case <-finished:
		// OK, колбэк завершился
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Callback did not finish in time")
	}

	elapsed := time.Since(start)
	if elapsed < 100*time.Millisecond {
		t.Errorf("Expected StopAll to block at least 100ms, got: %v", elapsed)
	}
}

// TestStartRequestTriggersSoftTimeout_Flaky проверяет срабатывание софт таймаута
func TestStartRequestTriggersSoftTimeout_Flaky(t *testing.T) {
	manager := NewRequestLifecycleManager[string](nil)
	id := newID("trigger-soft")

	triggered := make(chan TimeoutType, 1)
	err := manager.StartRequest(id, 10*time.Millisecond, time.Second, func(_ ID[string], tt TimeoutType) {
		triggered <- tt
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	select {
	case tt := <-triggered:
		if tt != SoftTimeout {
			t.Errorf("Expected SoftTimeout, got: %v", tt)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected soft timeout to be triggered")
	}
}

// TestStartRequestTriggersMaximumTimeout_Flaky проверяет срабатывание максимального таймаута
func TestStartRequestTriggersMaximumTimeout_Flaky(t *testing.T) {
	manager := NewRequestLifecycleManager[string](nil)
	id := newID("trigger-maximum")

	triggered := make(chan TimeoutType, 1)
	err := manager.StartRequest(id, 20*time.Millisecond, 20*time.Millisecond, func(_ ID[string], tt TimeoutType) {
		triggered <- tt
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var result TimeoutType
	select {
	case tt := <-triggered:
		result = tt
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Expected maximum timeout to be triggered")
	}

	if result != MaximumTimeout {
		t.Errorf("Expected MaximumTimeout, got: %v", result)
	}
}
