package protocol

import (
	"encoding/json"
	"testing"
)

func TestValidateValidNotification(t *testing.T) {
	n := jsonRPCNotification{
		JSONRPC: JSONRPCVersion,
		Method:  "log.message",
		Params:  map[string]string{"msg": "hi"},
	}
	if err := n.validate(); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateInvalidVersionNotification(t *testing.T) {
	n := jsonRPCNotification{
		JSONRPC: "1.0",
		Method:  "log",
	}
	err := n.validate()
	if err == nil || err.Error() != `invalid JSON-RPC version: expected "2.0", got "1.0"` {
		t.Errorf("Expected version error, got: %v", err)
	}
}

func TestValidateEmptyMethod(t *testing.T) {
	n := jsonRPCNotification{
		JSONRPC: JSONRPCVersion,
		Method:  "  ",
	}
	err := n.validate()
	if err == nil || err.Error() != "method cannot be empty" {
		t.Errorf("Expected empty method error, got: %v", err)
	}
}

func TestValidateReservedPrefix(t *testing.T) {
	n := jsonRPCNotification{
		JSONRPC: JSONRPCVersion,
		Method:  "rpc.ping",
	}
	err := n.validate()
	if err == nil || err.Error() != "method names starting with 'rpc.' are reserved" {
		t.Errorf("Expected reserved method error, got: %v", err)
	}
}

func TestUnmarshalJSONValidNotification(t *testing.T) {
	data := []byte(`{"jsonrpc":"2.0","method":"log","params":{"msg":"hi"}}`)
	var n jsonRPCNotification
	err := json.Unmarshal(data, &n)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if n.Method != "log" {
		t.Errorf("Expected method 'log', got: %s", n.Method)
	}
}

func TestUnmarshalJSONEmptyDataNotification(t *testing.T) {
	var n jsonRPCNotification
	err := n.UnmarshalJSON([]byte{})
	if err == nil || err != ErrEmptyJSONData {
		t.Errorf("Expected ErrEmptyJSONData, got: %v", err)
	}
}

func TestUnmarshalJSONInvalidJSONNotification(t *testing.T) {
	var n jsonRPCNotification
	err := n.UnmarshalJSON([]byte(`{invalid}`))
	if err == nil {
		t.Errorf("Expected error on invalid JSON")
	}
}

func TestUnmarshalJSONValidationError(t *testing.T) {
	// Используем неверную версию JSON-RPC, чтобы вызвать ошибку валидации
	var n jsonRPCNotification
	err := n.UnmarshalJSON([]byte(`{"jsonrpc":"1.0","method":"test"}`))
	if err == nil {
		t.Errorf("Expected validation error, got nil")
	}
	if _, ok := err.(*ValidationError); !ok {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestGetSetMethodNotification(t *testing.T) {
	n := NewNotification("start", nil)
	if n.GetMethod() != "start" {
		t.Errorf("Expected 'start', got: %s", n.GetMethod())
	}
	n.SetMethod("stop")
	if n.GetMethod() != "stop" {
		t.Errorf("Expected 'stop', got: %s", n.GetMethod())
	}
}

func TestGetSetParamsNotification(t *testing.T) {
	n := NewNotification("op", nil)
	if n.GetParams() != nil {
		t.Errorf("Expected nil params")
	}
	n.SetParams(map[string]string{"key": "value"})
	params, ok := n.GetParams().(map[string]string)
	if !ok || params["key"] != "value" {
		t.Errorf("Expected map with 'key':'value', got: %v", n.GetParams())
	}
}

func TestNewNotificationReturnsNotification(t *testing.T) {
	n := NewNotification("example.method", "param")
	if n.GetMethod() != "example.method" {
		t.Errorf("Expected method 'example.method', got: %s", n.GetMethod())
	}
	if n.GetParams() != "param" {
		t.Errorf("Expected param 'param', got: %v", n.GetParams())
	}
}
