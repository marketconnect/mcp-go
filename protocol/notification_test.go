package protocol

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewNotification_Basic(t *testing.T) {
	params := map[string]interface{}{"key": "value"}
	n := NewNotification("log.message", params)

	if n.JSONRPC != JSONRPCVersion {
		t.Errorf("expected JSONRPC version %q, got %q", JSONRPCVersion, n.JSONRPC)
	}

	if n.Method != "log.message" {
		t.Errorf("expected method to be 'log.message', got %q", n.Method)
	}

	if n.Params == nil {
		t.Error("expected params to be set")
	}
}

func TestNotification_Validate_Valid(t *testing.T) {
	valid := JSONRPCNotification{
		JSONRPC: JSONRPCVersion,
		Method:  "echo",
		Params:  nil,
	}

	if err := valid.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestNotification_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name string
		n    JSONRPCNotification
		want string
	}{
		{
			name: "invalid version",
			n: JSONRPCNotification{
				JSONRPC: "1.0",
				Method:  "test",
			},
			want: "invalid JSON-RPC version",
		},
		{
			name: "empty method",
			n: JSONRPCNotification{
				JSONRPC: JSONRPCVersion,
				Method:  "",
			},
			want: "method cannot be empty",
		},
		{
			name: "reserved method name",
			n: JSONRPCNotification{
				JSONRPC: JSONRPCVersion,
				Method:  "rpc.log",
			},
			want: "method names starting with 'rpc.'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.n.Validate()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("expected error to contain %q, got %q", tt.want, err.Error())
			}
		})
	}
}

func TestNotification_JSON_Serialization(t *testing.T) {
	n := NewNotification("notify", map[string]string{"msg": "hi"})
	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Проверяем, что "id" действительно отсутствует
	if bytes.Contains(data, []byte(`"id"`)) {
		t.Error(`notification JSON must not contain "id" field`)
	}

	// Десериализация обратно
	var decoded JSONRPCNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Method != "notify" {
		t.Errorf("expected method 'notify', got %q", decoded.Method)
	}
}
