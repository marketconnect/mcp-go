package protocol

import "testing"

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
