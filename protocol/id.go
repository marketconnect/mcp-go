package protocol

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
)

// https://spec.modelcontextprotocol.io/specification/2025-03-26/basic/#requests
// IDConstraint is a constraint that can be used to define the type of an ID.
// It can be a string or integer. Unlike base JSON-RPC, the ID MUST NOT be null.

type IDConstraint interface {
	~string | ~int | ~int64
}

// IDType is a generic type that can hold any type that satisfies the IDConstraint.
type IDType[T IDConstraint] struct {
	Value T
}

// NewID wraps a raw ID value (string or integer) into an IDType.
//
// This function is useful when you have an existing raw ID and want to create an IDType
// for use in requests or responses.
//
// Note: When using NewID, it is your responsibility to ensure that the provided ID
// is unique within the same session, as required by the MCP protocol:
// "The request ID MUST NOT have been previously used by the requestor within the same session."
// See https://spec.modelcontextprotocol.io/specification/2025-03-26/basic/#requests
//
// For automatic generation of unique IDs, consider using NextIntID or NextStringID.
//
// Example:
//
//	id := protocol.NewID(42)
func NewID[T IDConstraint](value T) IDType[T] {
	return IDType[T]{Value: value}
}

// idCounter is a global atomic counter used to generate unique IDs.
//
// It is used internally by NextIntID and NextStringID functions
// to ensure ID uniqueness across concurrent operations.
var idCounter int64

// NextIntID generates a new unique int64-based ID.
//
// It is thread-safe and guarantees unique IDs in concurrent environments.
//
// Example:
//
//	id := protocol.NextIntID()
func NextIntID() IDType[int64] {
	id := atomic.AddInt64(&idCounter, 1)
	return NewID(id)
}

// NextStringID generates a new unique string-based ID.
//
// The ID is formatted as "req-{counter}" where counter is an atomic incrementing number.
// It is thread-safe and guarantees unique IDs in concurrent environments.
//
// Example:
//
//	id := protocol.NextStringID()
//	fmt.Println(id.String()) // Output: req-1, req-2, etc.
func NextStringID() IDType[string] {

	id := atomic.AddInt64(&idCounter, 1)
	return NewID(fmt.Sprintf("req-%d", id))
}

// String returns the string representation of the IDType.
//
// Example:
//
//	id := protocol.NewID(42)
//	fmt.Println(id.String()) // Output: 42
func (id IDType[T]) String() string {
	return fmt.Sprintf("%v", id.Value)
}

// IsEmpty checks if the IDType contains the zero value of its underlying type.
//
// Example:
//
//	id := protocol.NewID("")
//	if id.IsEmpty() {
//	    log.Println("ID is empty")
//	}
func (id IDType[T]) IsEmpty() bool {
	var zero T
	return id.Value == zero
}

// Equal compares two IDType values for equality.
//
// Example:
//
//	id1 := protocol.NewID(42)
//	id2 := protocol.NewID(42)
//	fmt.Println(id1.Equal(id2)) // Output: true
func (id IDType[T]) Equal(other IDType[T]) bool {
	return id.Value == other.Value
}

// MarshalJSON implements the json.Marshaler interface.
//
// It ensures that the IDType is serialized as its underlying primitive value
// (string or integer), not as a nested object.
//
// Example output:
//
//	{"id": 42}
func (id IDType[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.Value)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// It parses the JSON value into the IDType and validates that it is not empty.
//
// Example:
//
//	var id protocol.IDType[int]
//	err := json.Unmarshal([]byte("42"), &id)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (id *IDType[T]) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &id.Value); err != nil {
		return &InvalidIDError{Err: err}
	}
	if id.IsEmpty() {
		return ErrEmptyRequestID
	}
	return nil
}
