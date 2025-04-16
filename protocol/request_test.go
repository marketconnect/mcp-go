package protocol_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/marketconnect/mcp-go/protocol"
)

func TestNewRequest(t *testing.T) {
	req := protocol.NewRequest("sum", 123, map[string]interface{}{"a": 1, "b": 2})

	if req.JSONRPC != protocol.JSONRPCVersion {
		t.Errorf("expected JSONRPC version %q, got %q", protocol.JSONRPCVersion, req.JSONRPC)
	}

	if req.Method != "sum" {
		t.Errorf("expected method 'sum', got %q", req.Method)
	}

	if req.ID.Value != 123 {
		t.Errorf("expected ID 123, got %v", req.ID.Value)
	}

	params, ok := req.Params.(map[string]interface{})
	if !ok || params["a"] != 1 || params["b"] != 2 {
		t.Errorf("unexpected params: %#v", req.Params)
	}
}

func TestNewRequestWithID(t *testing.T) {
	id := protocol.NewID(456)
	req := protocol.NewRequestWithID("multiply", id, map[string]interface{}{"x": 3, "y": 4})

	if req.ID != id {
		t.Errorf("expected ID %v, got %v", id, req.ID)
	}
	if req.Method != "multiply" {
		t.Errorf("expected method 'multiply', got %q", req.Method)
	}
}

func TestRequestValidate(t *testing.T) {
	validReq := protocol.NewRequest("sum", 123, nil)
	if err := validReq.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	cases := []struct {
		name    string
		request protocol.JSONRPCRequest[int]
		wantErr string
	}{
		{
			name:    "invalid JSONRPC version",
			request: protocol.JSONRPCRequest[int]{JSONRPC: "1.0", Method: "sum", ID: protocol.NewID(1)},
			wantErr: "invalid JSON-RPC version",
		},
		{
			name:    "empty method name",
			request: protocol.JSONRPCRequest[int]{JSONRPC: protocol.JSONRPCVersion, Method: "", ID: protocol.NewID(1)},
			wantErr: "method name cannot be empty",
		},

		{
			name:    "reserved method name",
			request: protocol.JSONRPCRequest[int]{JSONRPC: protocol.JSONRPCVersion, Method: "rpc.call", ID: protocol.NewID(1)},
			wantErr: "method names starting with 'rpc.' are reserved",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.request.Validate()
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("expected error containing %q, got %v", c.wantErr, err)
			}
		})
	}
}

func TestRequestUnmarshalJSON(t *testing.T) {
	jsonData := `{
		"jsonrpc": "2.0",
		"method": "sum",
		"id": 42,
		"params": {"a": 1, "b": 2}
	}`

	var req protocol.JSONRPCRequest[int]
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Fatalf("unexpected error unmarshalling JSON: %v", err)
	}

	if req.JSONRPC != protocol.JSONRPCVersion {
		t.Errorf("expected JSONRPC %q, got %q", protocol.JSONRPCVersion, req.JSONRPC)
	}

	if req.Method != "sum" {
		t.Errorf("expected method 'sum', got %q", req.Method)
	}

	if req.ID.Value != 42 {
		t.Errorf("expected ID 42, got %v", req.ID.Value)
	}
}

func TestRequestMarshalUnmarshalJSON(t *testing.T) {
	original := protocol.NewRequest("sum", 123, map[string]interface{}{"a": 1, "b": 2})

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("unexpected error marshalling JSON: %v", err)
	}

	var decoded protocol.JSONRPCRequest[int]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected error unmarshalling JSON: %v", err)
	}

	if decoded.Method != original.Method {
		t.Errorf("expected method %q, got %q", original.Method, decoded.Method)
	}

	if decoded.ID != original.ID {
		t.Errorf("expected ID %v, got %v", original.ID, decoded.ID)
	}
}

func TestRequestGetID(t *testing.T) {
	req := protocol.NewRequest("sum", 789, nil)
	if got := req.GetID(); got != 789 {
		t.Errorf("expected ID 789, got %v", got)
	}
}
