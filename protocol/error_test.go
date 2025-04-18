package protocol

import (
	"errors"
	"fmt"
	"testing"
)

func TestValidationError(t *testing.T) {
	err := &ValidationError{Reason: "missing field"}
	if err.Error() != "missing field" {
		t.Errorf("Expected error message 'missing field', got: %s", err.Error())
	}
}

func TestInvalidIDError(t *testing.T) {
	wrapped := errors.New("unexpected type")
	err := &InvalidIDError{Err: wrapped}
	expected := fmt.Sprintf("%v: %v", ErrInvalidID, wrapped)

	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got: %s", expected, err.Error())
	}
}

func TestInvalidIDErrorUnwrap(t *testing.T) {
	innerErr := errors.New("test error")
	err := &InvalidIDError{Err: innerErr}

	unwrapped := err.Unwrap()
	if unwrapped != ErrInvalidID {
		t.Errorf("Expected Unwrap to return ErrInvalidID, got %v", unwrapped)
	}

	// Проверяем, что errors.Is корректно работает с Unwrap
	if !errors.Is(err, ErrInvalidID) {
		t.Error("Expected errors.Is to return true for ErrInvalidID")
	}
}

func TestInvalidIDErrorIs(t *testing.T) {
	err := &InvalidIDError{Err: errors.New("test error")}

	// Проверяем метод Is напрямую
	if !err.Is(ErrInvalidID) {
		t.Error("Expected Is to return true for ErrInvalidID")
	}

	if err.Is(errors.New("random error")) {
		t.Error("Expected Is to return false for other errors")
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("field %s is required", "name")
	expected := "field name is required"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got: %s", expected, err.Error())
	}
}

func TestNewInvalidIDError(t *testing.T) {
	err := NewInvalidIDError("id %s not valid", "abc")
	expected := fmt.Sprintf("%v: id abc not valid", ErrInvalidID)
	if err.Error() != expected {
		t.Errorf("Expected '%s', got: %s", expected, err.Error())
	}
}
