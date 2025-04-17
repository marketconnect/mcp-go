package protocol

import (
	"fmt"
	"strings"
)

// JSONRPCNotification represents a JSON-RPC 2.0 Notification.
// A notification is a request without an "id" field.
type JSONRPCNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// NewNotification creates a new JSON-RPC notification.
//
// Example:
//     note := protocol.NewNotification("log", map[string]interface{}{"msg": "hello"})

func NewNotification(method string, params interface{}) JSONRPCNotification {
	return JSONRPCNotification{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  params,
	}
}

// Validate checks the Notification for correctness.
func (n JSONRPCNotification) Validate() error {
	if n.JSONRPC != JSONRPCVersion {
		return &ValidationError{Reason: fmt.Sprintf("invalid JSON-RPC version: expected %q, got %q", JSONRPCVersion, n.JSONRPC)}
	}
	if strings.TrimSpace(n.Method) == "" {
		return &ValidationError{Reason: "method cannot be empty"}
	}
	if strings.HasPrefix(n.Method, "rpc.") {
		return &ValidationError{Reason: "method names starting with 'rpc.' are reserved"}
	}
	return nil
}
