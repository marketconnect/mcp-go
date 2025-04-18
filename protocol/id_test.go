package protocol

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

func TestNewIDCreatesCorrectValue(t *testing.T) {
	IntegerID := newID(42)
	if IntegerID.Value != 42 {
		t.Errorf("Expected 42, got %v", IntegerID.Value)
	}

	StringID := newID("custom")
	if StringID.Value != "custom" {
		t.Errorf("Expected 'custom', got %v", StringID.Value)
	}
}

func TestIsEmptyDetectsZeroValues(t *testing.T) {
	EmptyIntID := ID[int64]{}
	if !EmptyIntID.isEmpty() {
		t.Errorf("Expected int64 zero to be empty")
	}

	NonEmptyIntID := ID[int64]{Value: 1}
	if NonEmptyIntID.isEmpty() {
		t.Errorf("Expected non-zero int64 to be not empty")
	}

	EmptyStringID := ID[string]{}
	if !EmptyStringID.isEmpty() {
		t.Errorf("Expected empty string to be empty")
	}

	NonEmptyStringID := ID[string]{Value: "abc"}
	if NonEmptyStringID.isEmpty() {
		t.Errorf("Expected non-empty string to be not empty")
	}
}

func TestNextIntIDGeneratesUniqueValues(t *testing.T) {
	Seen := make(map[int64]struct{})
	var Mutex sync.Mutex
	var WaitGroup sync.WaitGroup

	for i := 0; i < 10; i++ {
		WaitGroup.Add(1)
		go func() {
			defer WaitGroup.Done()
			for j := 0; j < 100; j++ {
				ID := NextIntID()
				Mutex.Lock()
				if _, Exists := Seen[ID.Value]; Exists {
					t.Errorf("Duplicate ID generated: %d", ID.Value)
				}
				Seen[ID.Value] = struct{}{}
				Mutex.Unlock()
			}
		}()
	}
	WaitGroup.Wait()
}

func TestNextStringIDReturnsPrefixedString(t *testing.T) {
	ID := NextStringID()
	if !strings.HasPrefix(ID.Value, "req-") {
		t.Errorf("Expected string ID to start with 'req-', got: %s", ID.Value)
	}
}

func TestMarshalJSONReturnsPrimitiveValue(t *testing.T) {
	id := ID[int64]{Value: 123}
	Data, Err := json.Marshal(id)
	if Err != nil {
		t.Fatalf("Failed to marshal ID: %v", Err)
	}
	if string(Data) != "123" {
		t.Errorf("Expected '123', got %s", string(Data))
	}

	StringID := ID[string]{Value: "test"}
	Data, Err = json.Marshal(StringID)
	if Err != nil {
		t.Fatalf("Failed to marshal string ID: %v", Err)
	}
	if string(Data) != `"test"` {
		t.Errorf(`Expected '"test"', got %s`, string(Data))
	}
}

func TestUnmarshalJSONWithValidInput(t *testing.T) {
	var IntID ID[int64]
	Err := json.Unmarshal([]byte("456"), &IntID)
	if Err != nil {
		t.Fatalf("Unmarshal failed: %v", Err)
	}
	if IntID.Value != 456 {
		t.Errorf("Expected 456, got %v", IntID.Value)
	}

	var StrID ID[string]
	Err = json.Unmarshal([]byte(`"req-789"`), &StrID)
	if Err != nil {
		t.Fatalf("Unmarshal failed: %v", Err)
	}
	if StrID.Value != "req-789" {
		t.Errorf("Expected 'req-789', got %v", StrID.Value)
	}
}

func TestUnmarshalJSONRejectsEmptyValues(t *testing.T) {
	var EmptyStrID ID[string]
	Err := json.Unmarshal([]byte(`""`), &EmptyStrID)
	if Err == nil {
		t.Fatalf("Expected error for empty string, got nil")
	}
	if Err != ErrEmptyRequestID {
		t.Errorf("Expected ErrEmptyRequestID, got %v", Err)
	}

	var EmptyIntID ID[int64]
	Err = json.Unmarshal([]byte("0"), &EmptyIntID)
	if Err == nil {
		t.Fatalf("Expected error for zero int, got nil")
	}
	if Err != ErrEmptyRequestID {
		t.Errorf("Expected ErrEmptyRequestID, got %v", Err)
	}
}

func TestUnmarshalJSONReturnsInvalidIDErrorForGarbage(t *testing.T) {
	var ID ID[int64]
	Err := json.Unmarshal([]byte(`{}`), &ID)
	if Err == nil {
		t.Fatalf("Expected error for invalid input, got nil")
	}
	_, IsTyped := Err.(*InvalidIDError)
	if !IsTyped {
		t.Errorf("Expected InvalidIDError, got %T", Err)
	}
}

func TestNextIntID(t *testing.T) {
	id1 := NextIntID()
	id2 := NextIntID()

	// Проверяем, что ID уникальные и увеличиваются
	if id1.Value >= id2.Value {
		t.Errorf("Expected NextIntID to generate increasing values, got %d, then %d", id1.Value, id2.Value)
	}
}

func TestNextStringID(t *testing.T) {
	id1 := NextStringID()
	id2 := NextStringID()

	// Проверяем, что ID имеют ожидаемую форму
	if id1.Value == "" || id1.Value[0:4] != "req-" {
		t.Errorf("Expected NextStringID to start with 'req-', got %s", id1.Value)
	}

	// Проверяем, что ID уникальные
	if id1.Value == id2.Value {
		t.Errorf("Expected NextStringID to generate unique values, got %s twice", id1.Value)
	}
}

func TestMarshalJSON(t *testing.T) {
	// ID с integer
	intID := newID(42)
	intData, err := intID.MarshalJSON()
	if err != nil {
		t.Fatalf("Error marshaling int ID: %v", err)
	}
	if string(intData) != "42" {
		t.Errorf("Expected marshaled int ID to be '42', got %s", string(intData))
	}

	// ID со string
	strID := newID("test-id")
	strData, err := strID.MarshalJSON()
	if err != nil {
		t.Fatalf("Error marshaling string ID: %v", err)
	}
	if string(strData) != `"test-id"` {
		t.Errorf("Expected marshaled string ID to be '\"test-id\"', got %s", string(strData))
	}

	// Проверка, что MarshalJSON корректно сериализует ID как примитивное значение
	type Container struct {
		ID ID[string] `json:"id"`
	}

	container := Container{ID: newID("serialized-id")}
	containerData, err := json.Marshal(container)
	if err != nil {
		t.Fatalf("Error marshaling container: %v", err)
	}

	expected := `{"id":"serialized-id"}`
	if string(containerData) != expected {
		t.Errorf("Expected marshaled container to be '%s', got '%s'", expected, string(containerData))
	}
}
