package protocol_test

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/marketconnect/mcp-go/protocol"
)

func TestNewID(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"Int ID", 123, 123},
		{"String ID", "abc", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.input.(type) {
			case int:
				id := protocol.NewID(v)
				if id.Value != tt.expected {
					t.Errorf("expected ID value %v, got %v", tt.expected, id.Value)
				}
			case string:
				id := protocol.NewID(v)
				if id.Value != tt.expected {
					t.Errorf("expected ID value %q, got %v", tt.expected, id.Value)
				}
			}
		})
	}
}

func TestNextIntID_Uniqueness(t *testing.T) {
	id1 := protocol.NextIntID()
	id2 := protocol.NextIntID()

	if id1.Equal(id2) {
		t.Errorf("expected unique IDs, but got duplicates for %v and %v", id1, id2)
	}
}

func TestNextIntID_Concurrency(t *testing.T) {
	const goroutines = 100
	ids := make(chan int64, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			id := protocol.NextIntID()
			ids <- id.Value
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[int64]struct{})
	for id := range ids {
		if _, exists := seen[id]; exists {
			t.Errorf("duplicate ID generated: %v", id)
		}
		seen[id] = struct{}{}
	}
}

func TestNextStringID_Uniqueness(t *testing.T) {
	id1 := protocol.NextStringID()
	id2 := protocol.NextStringID()

	if id1.Equal(id2) {
		t.Error("expected unique string IDs, but got duplicates")
	}
}

func TestIDType_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		id       interface{}
		expected bool
	}{
		{"Empty int ID", protocol.NewID(0), true},
		{"Empty string ID", protocol.NewID(""), true},
		{"Non-empty int ID", protocol.NewID(42), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.id.(type) {
			case protocol.IDType[int]:
				if v.IsEmpty() != tt.expected {
					t.Errorf("expected IsEmpty to be %v, got %v", tt.expected, v.IsEmpty())
				}
			case protocol.IDType[string]:
				if v.IsEmpty() != tt.expected {
					t.Errorf("expected IsEmpty to be %v, got %v", tt.expected, v.IsEmpty())
				}
			}
		})
	}
}
func TestIDType_Equal(t *testing.T) {
	tests := []struct {
		name     string
		id1      interface{}
		id2      interface{}
		expected bool
	}{
		{"Equal int IDs", protocol.NewID(42), protocol.NewID(42), true},
		{"Different int IDs", protocol.NewID(42), protocol.NewID(43), false},
		{"Equal string IDs", protocol.NewID("foo"), protocol.NewID("foo"), true},
		{"Different string IDs", protocol.NewID("foo"), protocol.NewID("bar"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v1 := tt.id1.(type) {
			case protocol.IDType[int]:
				v2 := tt.id2.(protocol.IDType[int])
				if v1.Equal(v2) != tt.expected {
					t.Errorf("expected Equal to be %v, got %v", tt.expected, v1.Equal(v2))
				}
			case protocol.IDType[string]:
				v2 := tt.id2.(protocol.IDType[string])
				if v1.Equal(v2) != tt.expected {
					t.Errorf("expected Equal to be %v, got %v", tt.expected, v1.Equal(v2))
				}
			}
		})
	}
}

func TestIDType_String(t *testing.T) {
	tests := []struct {
		name     string
		id       interface{}
		expected string
	}{
		{"Int ID to string", protocol.NewID(42), "42"},
		{"String ID to string", protocol.NewID("abc"), "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.id.(type) {
			case protocol.IDType[int]:
				if v.String() != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, v.String())
				}
			case protocol.IDType[string]:
				if v.String() != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, v.String())
				}
			}
		})
	}
}

func TestIDType_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		id       interface{}
		expected string
	}{
		{"Marshal int ID", protocol.NewID(42), "42"},
		{"Marshal string ID", protocol.NewID("abc"), `"abc"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err error
			switch v := tt.id.(type) {
			case protocol.IDType[int]:
				data, err = json.Marshal(v)
			case protocol.IDType[string]:
				data, err = json.Marshal(v)
			}
			if err != nil {
				t.Fatalf("failed to marshal ID: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("expected JSON %s, got %s", tt.expected, string(data))
			}
		})
	}
}
func TestIDType_UnmarshalJSON(t *testing.T) {
	var id protocol.IDType[int]
	data := []byte("42")
	if err := json.Unmarshal(data, &id); err != nil {
		t.Fatalf("failed to unmarshal ID: %v", err)
	}
	if id.Value != 42 {
		t.Errorf("expected ID value 42, got %v", id.Value)
	}

	var idStr protocol.IDType[string]
	dataStr := []byte(`"abc"`)
	if err := json.Unmarshal(dataStr, &idStr); err != nil {
		t.Fatalf("failed to unmarshal string ID: %v", err)
	}
	if idStr.Value != "abc" {
		t.Errorf(`expected ID value "abc", got %v`, idStr.Value)
	}
}

func TestIDType_UnmarshalJSON_EmptyID(t *testing.T) {
	var id protocol.IDType[int]
	data := []byte("0")
	err := json.Unmarshal(data, &id)

	if err == nil {
		t.Fatal("expected error for empty int ID, got nil")
	}

	if !errors.Is(err, protocol.ErrEmptyRequestID) {
		t.Errorf("expected ErrEmptyRequestID, got %v", err)
	}

	var idStr protocol.IDType[string]
	dataStr := []byte(`""`)
	err = json.Unmarshal(dataStr, &idStr)

	if err == nil {
		t.Fatal("expected error for empty string ID, got nil")
	}

	if !errors.Is(err, protocol.ErrEmptyRequestID) {
		t.Errorf("expected ErrEmptyRequestID, got %v", err)
	}
}
func TestIDType_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var id protocol.IDType[int]
	data := []byte(`{}`)

	err := json.Unmarshal(data, &id)

	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	var invalidIDErr *protocol.InvalidIDError
	if !errors.As(err, &invalidIDErr) {
		t.Errorf("expected InvalidIDError, got %v", err)
	}
}

func BenchmarkNextIntID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = protocol.NextIntID()
	}
}

func BenchmarkNextStringID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = protocol.NextStringID()
	}
}
