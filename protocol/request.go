package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
)

const JSONRPCVersion = "2.0"

type jsonRPCRequest[T IDConstraint] struct {
	// A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JSONRPC string `json:"jsonrpc"`
	// A String containing the name of the method to be invoked.
	// Method names that begin with the word rpc followed by a period character (U+002E or ASCII 46) are reserved for rpc-internal methods
	// and extensions and MUST NOT be used for anything else.
	Method string `json:"method"`
	// A Structured value that holds the parameter values to be used during the invocation of the method.
	// This member MAY be omitted.
	Params interface{} `json:"params,omitempty"` // optional parameters

	ID ID[T] `json:"id"`
}

// validate checks if the JSON-RPC request is valid.
func (r jsonRPCRequest[T]) validate() error {
	if r.JSONRPC != JSONRPCVersion {
		return &ValidationError{Reason: fmt.Sprintf("invalid JSON-RPC version: expected %q, got %q", JSONRPCVersion, r.JSONRPC)}
	}

	if len(strings.TrimSpace(r.Method)) == 0 {
		return &ValidationError{Reason: "method name cannot be empty or whitespace"}
	}

	if r.ID.isEmpty() {
		return &ValidationError{Reason: "id must not be empty"}
	}

	// https://www.jsonrpc.org/specification#request_object
	// Method names that begin with the word rpc followed by
	// a period character (U+002E or ASCII 46) are reserved
	// for rpc-internal methods and extensions and MUST NOT be used for anything else.
	if strings.HasPrefix(r.Method, "rpc.") {
		return &ValidationError{Reason: fmt.Sprintf("method names starting with 'rpc.' are reserved, got: %q", r.Method)}
	}

	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// It parses the JSON-RPC request from JSON data and validates its correctness.
// If the input is invalid or violates the JSON-RPC specification, an error is returned.
//
// This method automatically calls Validate() after unmarshaling.
//
// Example:
//
//	var req protocol.JSONRPCRequest[int]
//	err := json.Unmarshal(data, &req)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (r *jsonRPCRequest[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return ErrEmptyJSONData
	}

	type requestNoMethods jsonRPCRequest[T]
	aux := &struct {
		requestNoMethods
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	temp := jsonRPCRequest[T](aux.requestNoMethods)
	if err := temp.validate(); err != nil {
		return err
	}
	*r = temp
	return nil

}

func (r *jsonRPCRequest[T]) GetID() interface{} {
	return r.ID.Value
}

func (r *jsonRPCRequest[T]) SetID(v interface{}) error {
	val, ok := v.(T)
	if !ok {
		return ErrInvalidID
	}
	r.ID = ID[T]{Value: val}
	return nil
}

func (r *jsonRPCRequest[T]) GetMethod() string {
	return r.Method
}

func (r *jsonRPCRequest[T]) SetMethod(method string) {
	r.Method = method
}

func (r *jsonRPCRequest[T]) GetParams() interface{} {
	return r.Params
}

func (r *jsonRPCRequest[T]) SetParams(params interface{}) {
	r.Params = params
}

// NewRequest creates a new JSON-RPC request object with the given method, params, and ID.
//
// Example:
//
//	req := protocol.NewRequest("myMethod", map[string]any{"foo": "bar"}, protocol.NextIntID())
func NewRequest[T IDConstraint](method string, params interface{}, id ID[T]) Request {

	return &jsonRPCRequest[T]{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  params,
		ID:      id,
	}
}

type Request interface {
	GetID() any
	SetID(any) error
	GetMethod() string
	SetMethod(string)
	GetParams() interface{}
	SetParams(interface{})
}
