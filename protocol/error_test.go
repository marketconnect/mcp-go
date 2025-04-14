package protocol_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/marketconnect/mcp-go/protocol"
)

func TestErrorVariables(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrInvalidID", protocol.ErrInvalidID, "invalid ID"},
		{"ErrEmptyRequestID", protocol.ErrEmptyRequestID, "request ID cannot be empty"},
		{"ErrSoftTimeoutNotPositive", protocol.ErrSoftTimeoutNotPositive, "soft timeout must be greater than zero"},
		{"ErrMaximumTimeoutNotPositive", protocol.ErrMaximumTimeoutNotPositive, "maximum timeout must be greater than zero"},
		{"ErrSoftTimeoutExceedsMaximum", protocol.ErrSoftTimeoutExceedsMaximum, "soft timeout exceeds or equals maximum timeout"},
		{"ErrDuplicateRequestID", protocol.ErrDuplicateRequestID, "request ID already used in this session"},
		{"ErrRequestNotFound", protocol.ErrRequestNotFound, "request not found"},
		{"ErrCallbackNil", protocol.ErrCallbackNil, "callback must not be nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("error message = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	reason := "missing field"
	err := &protocol.ValidationError{Reason: reason}

	if err.Error() != reason {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), reason)
	}

	// Test via interface
	var e error = err
	if e.Error() != reason {
		t.Errorf("interface error message = %q, want %q", e.Error(), reason)
	}
}

func TestInvalidIDError(t *testing.T) {
	baseErr := errors.New("unexpected format")
	err := &protocol.InvalidIDError{Err: baseErr}

	// Check error string
	want := fmt.Sprintf("%v: %v", protocol.ErrInvalidID, baseErr)
	if err.Error() != want {
		t.Errorf("InvalidIDError.Error() = %q, want %q", err.Error(), want)
	}

	// Check Unwrap
	if !errors.Is(err, protocol.ErrInvalidID) {
		t.Errorf("errors.Is failed: expected to unwrap to ErrInvalidID")
	}

	// Check errors.As
	var target *protocol.InvalidIDError
	if !errors.As(err, &target) {
		t.Errorf("errors.As failed: expected target *InvalidIDError")
	}

	if target.Err != baseErr {
		t.Errorf("wrapped error mismatch: got %v, want %v", target.Err, baseErr)
	}
}

func TestNewValidationError(t *testing.T) {
	format := "invalid field: %s"
	field := "method"
	err := protocol.NewValidationError(format, field)

	want := fmt.Sprintf(format, field)
	if err.Error() != want {
		t.Errorf("NewValidationError() = %q, want %q", err.Error(), want)
	}
}

func TestNewInvalidIDError(t *testing.T) {
	format := "invalid ID format: %s"
	detail := "non-numeric string"
	err := protocol.NewInvalidIDError(format, detail)

	want := fmt.Sprintf("%v: %v", protocol.ErrInvalidID, fmt.Sprintf(format, detail))
	if err.Error() != want {
		t.Errorf("NewInvalidIDError() = %q, want %q", err.Error(), want)
	}

	if !errors.Is(err, protocol.ErrInvalidID) {
		t.Errorf("errors.Is failed: expected to unwrap to ErrInvalidID")
	}
}
