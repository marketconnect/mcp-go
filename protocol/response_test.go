package protocol_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/marketconnect/mcp-go/protocol"
)

type SampleResult struct {
	Value string `json:"value"`
}

func TestNewSuccessResponse(t *testing.T) {
	id := protocol.NewID("req-1")
	result := &SampleResult{Value: "OK"}

	resp := protocol.NewSuccessResponse(id.Value, result)

	if resp.JSONRPC != protocol.JSONRPCVersion {
		t.Errorf("expected JSONRPC version %q, got %q", protocol.JSONRPCVersion, resp.JSONRPC)
	}
	if resp.ID != id {
		t.Errorf("expected ID %v, got %v", id, resp.ID)
	}
	if resp.Result == nil || resp.Result.Value != "OK" {
		t.Errorf("unexpected result: %+v", resp.Result)
	}
	if resp.Error != nil {
		t.Errorf("expected no error, got %+v", resp.Error)
	}
}

func TestNewErrorResponse(t *testing.T) {

	resp := protocol.NewErrorResponse[string, any](protocol.NewID("req-1"), -32601, "not found", nil)

	if resp.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if resp.Error.Code != -32601 || resp.Error.Message != "Method not found" {
		t.Errorf("unexpected error content: %+v", resp.Error)
	}
	if resp.Result != nil {
		t.Errorf("expected no result, got %+v", resp.Result)
	}
}

func TestNewErrorResponseFromError(t *testing.T) {

	err := protocol.NewRPCError(-32602, "Invalid params", nil)

	resp := protocol.NewErrorResponseFromError[string, any](protocol.NewID("req-42"), err)

	if resp.Error != err {
		t.Errorf("expected error to be same instance")
	}
}

func TestValidate(t *testing.T) {
	type result struct{ Ok bool }
	validResp := protocol.NewSuccessResponse("id-1", &result{Ok: true})
	if err := validResp.Validate(); err != nil {
		t.Errorf("valid response validation failed: %v", err)
	}

	cases := []struct {
		name    string
		resp    protocol.JSONRPCResponse[string, result]
		wantErr string
	}{
		{
			"Empty ID",
			protocol.JSONRPCResponse[string, result]{JSONRPC: protocol.JSONRPCVersion, Result: &result{true}},
			"response ID must not be empty",
		},
		{
			"Missing result and error",
			protocol.JSONRPCResponse[string, result]{JSONRPC: protocol.JSONRPCVersion, ID: protocol.NewID("id")},
			"response MUST contain either result or error",
		},
		{
			"Both result and error",
			protocol.JSONRPCResponse[string, result]{
				JSONRPC: protocol.JSONRPCVersion,
				ID:      protocol.NewID("id"),
				Result:  &result{true},
				Error:   protocol.NewRPCError(-1, "fail", nil),
			},
			"response MUST NOT contain both result and error",
		},
		{
			"Missing error message",
			protocol.JSONRPCResponse[string, result]{
				JSONRPC: protocol.JSONRPCVersion,
				ID:      protocol.NewID("id"),
				Error:   &protocol.RPCError{Code: -1},
			},
			"error must contain non-empty message",
		},
		{
			"Invalid JSONRPC version",
			protocol.JSONRPCResponse[string, result]{
				JSONRPC: "1.0",
				ID:      protocol.NewID("id"),
				Result:  &result{true},
			},
			"invalid JSON-RPC version",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.resp.Validate()
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("expected error containing %q, got %v", c.wantErr, err)
			}
		})
	}
}

func TestIsSuccessAndError(t *testing.T) {
	respOK := protocol.NewSuccessResponse("id-ok", &SampleResult{Value: "👍"})
	if !respOK.IsSuccess() || respOK.IsError() {
		t.Error("expected success response to be IsSuccess=true, IsError=false")
	}

	respFail := protocol.NewErrorResponse[string, any](
		protocol.NewID("id-err"),
		-32001,
		"Something broke",
		nil,
	)
	if respFail.IsSuccess() || !respFail.IsError() {
		t.Error("expected error response to be IsError=true, IsSuccess=false")
	}
}

func TestGetID(t *testing.T) {
	resp := protocol.NewSuccessResponse("get-id", &SampleResult{Value: "OK"})
	if got := resp.GetID(); got != "get-id" {
		t.Errorf("expected ID 'get-id', got %v", got)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	raw := `{"jsonrpc":"2.0","id":"unmarshal-test","result":{"value":"decoded"}}`

	var decoded protocol.JSONRPCResponse[string, SampleResult]
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decoded.JSONRPC != "2.0" {
		t.Errorf("expected version 2.0, got %s", decoded.JSONRPC)
	}
	if decoded.ID.Value != "unmarshal-test" {
		t.Errorf("expected ID 'unmarshal-test', got %v", decoded.ID.Value)
	}
	if decoded.Result == nil || decoded.Result.Value != "decoded" {
		t.Errorf("unexpected result: %+v", decoded.Result)
	}
}
