package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
)

// jsonRPCNotification represents a JSON-RPC 2.0 Notification.
// A notification is a request without an "id" field.
type jsonRPCNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// validate checks the Notification for correctness.
func (n jsonRPCNotification) validate() error {
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

func (n *jsonRPCNotification) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return ErrEmptyJSONData
	}

	type notificationNoMethods jsonRPCNotification
	aux := &struct {
		notificationNoMethods
	}{}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	temp := jsonRPCNotification(aux.notificationNoMethods)
	if err := temp.validate(); err != nil {
		return err
	}
	*n = temp
	return nil
}

// GetMethod returns the method name of the notification.
func (n jsonRPCNotification) GetMethod() string {
	return n.Method
}

// SetMethod sets the method name of the notification.
func (n *jsonRPCNotification) SetMethod(method string) {
	n.Method = method
}

// GetParams returns the parameters of the notification.
func (n jsonRPCNotification) GetParams() interface{} {
	return n.Params
}

// SetParams sets the parameters of the notification.
func (n *jsonRPCNotification) SetParams(params interface{}) {
	n.Params = params
}

type Notification interface {
	GetMethod() string
	SetMethod(string)
	GetParams() interface{}
	SetParams(interface{})
}

func NewNotification(method string, params interface{}) Notification {
	return &jsonRPCNotification{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  params,
	}
}
