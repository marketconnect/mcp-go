package protocol

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestNewRequestConstructsProperly(t *testing.T) {
	Id := newID(1)
	Method := "doSomething"
	Params := map[string]interface{}{"x": 42}

	Req := NewRequest(Method, Params, Id)

	if Req.GetMethod() != Method {
		t.Errorf("Expected method '%s', got '%s'", Method, Req.GetMethod())
	}

	ParamsMap, Ok := Req.GetParams().(map[string]interface{})
	if !Ok {
		t.Fatalf("Expected Params to be a map, got %T", Req.GetParams())
	}
	XVal, Ok := ParamsMap["x"]
	if !Ok {
		t.Errorf("Expected param 'x' to exist")
	}
	XInt, Ok := XVal.(int)
	if !Ok || XInt != 42 {
		t.Errorf("Expected x = 42, got %v (type %T)", XVal, XVal)
	}

	if Req.GetID().(int) != 1 {
		t.Errorf("Expected ID = 1, got %v", Req.GetID())
	}
}

func TestSetID_ValidAndInvalid(t *testing.T) {
	Req := NewRequest("method", nil, newID(1)).(*jsonRPCRequest[int])

	Err := Req.SetID(100)
	if Err != nil {
		t.Errorf("Expected SetID to succeed, got error: %v", Err)
	}
	if Req.ID.Value != 100 {
		t.Errorf("Expected ID = 100, got %v", Req.ID.Value)
	}

	Err = Req.SetID("wrong-type")
	if !errors.Is(Err, ErrInvalidID) {
		t.Errorf("Expected ErrInvalidID, got %v", Err)
	}
}

func TestValidate_ValidRequest(t *testing.T) {
	Req := &jsonRPCRequest[int]{
		JSONRPC: JSONRPCVersion,
		Method:  "validMethod",
		Params:  nil,
		ID:      newID(42),
	}
	Err := Req.validate()
	if Err != nil {
		t.Errorf("Expected valid request, got error: %v", Err)
	}
}

func TestValidate_InvalidJSONRPCVersion(t *testing.T) {
	Req := &jsonRPCRequest[int]{
		JSONRPC: "1.0",
		Method:  "validMethod",
		ID:      newID(1),
	}
	Err := Req.validate()
	if Err == nil || Err.Error() != `invalid JSON-RPC version: expected "2.0", got "1.0"` {
		t.Errorf("Expected invalid JSON-RPC version error, got %v", Err)
	}
}

func TestValidate_EmptyMethod(t *testing.T) {
	Req := &jsonRPCRequest[int]{
		JSONRPC: JSONRPCVersion,
		Method:  "   ",
		ID:      newID(1),
	}
	Err := Req.validate()
	if Err == nil || Err.Error() != "method name cannot be empty or whitespace" {
		t.Errorf("Expected method name error, got %v", Err)
	}
}

func TestValidate_EmptyID(t *testing.T) {
	Req := &jsonRPCRequest[int]{
		JSONRPC: JSONRPCVersion,
		Method:  "test",
		ID:      newID(0),
	}
	Err := Req.validate()
	if Err == nil || Err.Error() != "id must not be empty" {
		t.Errorf("Expected empty ID error, got %v", Err)
	}
}

func TestValidate_ReservedMethodPrefix(t *testing.T) {
	Req := &jsonRPCRequest[int]{
		JSONRPC: JSONRPCVersion,
		Method:  "rpc.internal",
		ID:      newID(1),
	}
	Err := Req.validate()
	if Err == nil || Err.Error() != `method names starting with 'rpc.' are reserved, got: "rpc.internal"` {
		t.Errorf("Expected reserved prefix error, got %v", Err)
	}
}

func TestUnmarshalJSON_Valid(t *testing.T) {
	Data := `{"jsonrpc":"2.0","method":"compute","params":{"a":1},"id":999}`
	var Req jsonRPCRequest[int]
	Err := json.Unmarshal([]byte(Data), &Req)
	if Err != nil {
		t.Errorf("Expected valid unmarshal, got error: %v", Err)
	}
	if Req.ID.Value != 999 {
		t.Errorf("Expected ID = 999, got %v", Req.ID.Value)
	}
	if Req.Method != "compute" {
		t.Errorf("Expected method 'compute', got %v", Req.Method)
	}
}

func TestUnmarshalJSON_EmptyData(t *testing.T) {
	var Req jsonRPCRequest[int]
	Err := Req.UnmarshalJSON([]byte{})
	if !errors.Is(Err, ErrEmptyJSONData) {
		t.Errorf("Expected ErrEmptyJSONData, got %v", Err)
	}
}

func TestUnmarshalJSON_InvalidField(t *testing.T) {
	Data := `{"jsonrpc":"2.0","method":"  ","id":1}`
	var Req jsonRPCRequest[int]
	Err := json.Unmarshal([]byte(Data), &Req)
	if Err == nil {
		t.Errorf("Expected validation error due to empty method")
	}
}

func TestSetMethodAndSetParams(t *testing.T) {
	req := NewRequest("oldMethod", map[string]string{"old": "param"}, NextIntID()).(*jsonRPCRequest[int64])

	// Тест SetMethod
	req.SetMethod("newMethod")
	if req.GetMethod() != "newMethod" {
		t.Errorf("Expected method 'newMethod', got '%s'", req.GetMethod())
	}

	// Тест SetParams
	newParams := map[string]string{"new": "value"}
	req.SetParams(newParams)
	if !reflect.DeepEqual(req.GetParams(), newParams) {
		t.Errorf("Expected new params, got %v", req.GetParams())
	}
}

func TestUnmarshalJSONInvalidJSONStruct(t *testing.T) {
	// Создаем поврежденный JSON, который вызовет ошибку при разборе структуры
	invalidJSON := []byte(`{"jsonrpc":2.0,"method":"test","id":true}`)
	var req jsonRPCRequest[int]
	err := req.UnmarshalJSON(invalidJSON)
	if err == nil {
		t.Errorf("Expected error on invalid JSON structure, got nil")
	}
}
