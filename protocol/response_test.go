package protocol

import (
	"encoding/json"
	"testing"
)

func TestValidateValidResult(t *testing.T) {
	resp := jsonRPCResponse[string]{
		JSONRPC: JSONRPCVersion,
		ID:      ID[string]{Value: "123"},
		Result:  "success",
	}
	if err := resp.validate(); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateValidError(t *testing.T) {
	resp := jsonRPCResponse[string]{
		JSONRPC: JSONRPCVersion,
		ID:      ID[string]{Value: "123"},
		Error:   &RPCError{Code: -32601, Message: "Method not found"},
	}
	if err := resp.validate(); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateInvalidVersion(t *testing.T) {
	resp := jsonRPCResponse[string]{
		JSONRPC: "1.0",
		ID:      ID[string]{Value: "123"},
		Result:  "success",
	}
	err := resp.validate()
	if err == nil || err.Error() != "invalid JSON-RPC version: expected \"2.0\", got \"1.0\"" {
		t.Errorf("Expected version error, got: %v", err)
	}
}

func TestValidateEmptyID(t *testing.T) {
	resp := jsonRPCResponse[string]{
		JSONRPC: JSONRPCVersion,
		ID:      ID[string]{Value: ""},
		Result:  "success",
	}
	err := resp.validate()
	if err == nil || err.Error() != "response ID must not be empty" {
		t.Errorf("Expected empty ID error, got: %v", err)
	}
}

func TestValidateBothResultAndError(t *testing.T) {
	resp := jsonRPCResponse[string]{
		JSONRPC: JSONRPCVersion,
		ID:      ID[string]{Value: "123"},
		Result:  "success",
		Error:   &RPCError{Code: -32601, Message: "Method not found"},
	}
	err := resp.validate()
	if err == nil || err.Error() != "response MUST NOT contain both result and error" {
		t.Errorf("Expected both result and error error, got: %v", err)
	}
}

func TestValidateNeitherResultNorError(t *testing.T) {
	resp := jsonRPCResponse[string]{
		JSONRPC: JSONRPCVersion,
		ID:      ID[string]{Value: "123"},
	}
	err := resp.validate()
	if err == nil || err.Error() != "response MUST contain either result or error" {
		t.Errorf("Expected missing result and error error, got: %v", err)
	}
}

func TestValidateErrorWithoutMessage(t *testing.T) {
	resp := jsonRPCResponse[string]{
		JSONRPC: JSONRPCVersion,
		ID:      ID[string]{Value: "123"},
		Error:   &RPCError{Code: -32601, Message: ""},
	}
	err := resp.validate()
	if err == nil || err.Error() != "error must contain non-empty message" {
		t.Errorf("Expected error without message error, got: %v", err)
	}
}

func TestUnmarshalJSONValidResult(t *testing.T) {
	jsonData := []byte(`{"jsonrpc":"2.0","id":"123","result":"success"}`)
	var resp jsonRPCResponse[string]
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if resp.Result != "success" {
		t.Errorf("Expected result 'success', got: %v", resp.Result)
	}
}

func TestUnmarshalJSONValidError(t *testing.T) {
	jsonData := []byte(`{"jsonrpc":"2.0","id":"123","error":{"code":-32601,"message":"Method not found"}}`)
	var resp jsonRPCResponse[string]
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if resp.Error == nil || resp.Error.Message != "Method not found" {
		t.Errorf("Expected error message 'Method not found', got: %v", resp.Error)
	}
}

func TestUnmarshalJSONEmptyData(t *testing.T) {
	var resp jsonRPCResponse[string]
	err := resp.UnmarshalJSON([]byte{})
	if err == nil || err != ErrEmptyJSONData {
		t.Errorf("Expected ErrEmptyJSONData, got: %v", err)
	}
}

func TestUnmarshalJSONInvalidJSON(t *testing.T) {
	var resp jsonRPCResponse[string]
	err := resp.UnmarshalJSON([]byte(`invalid json`))
	if err == nil {
		t.Errorf("Expected JSON unmarshal error, got nil")
	}
}

func TestGetSetID(t *testing.T) {
	var resp jsonRPCResponse[string]
	err := resp.SetID("123")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if resp.GetID() != "123" {
		t.Errorf("Expected ID '123', got: %v", resp.GetID())
	}
}

func TestSetIDInvalidType(t *testing.T) {
	var resp jsonRPCResponse[string]
	err := resp.SetID(123) // int instead of string
	if err == nil {
		t.Errorf("Expected error for invalid ID type, got nil")
	}
}

func TestGetSetResult(t *testing.T) {
	var resp jsonRPCResponse[string]
	err := resp.SetResult("ok")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if resp.GetResult() != "ok" {
		t.Errorf("Expected result 'ok', got: %v", resp.GetResult())
	}
}

func TestGetSetError(t *testing.T) {
	var resp jsonRPCResponse[string]
	rpcErr := &RPCError{Code: -1, Message: "oops"}
	resp.SetError(rpcErr)
	if resp.GetError() != rpcErr {
		t.Errorf("Expected error object, got different")
	}
}

func TestHasResultTrue(t *testing.T) {
	resp := jsonRPCResponse[string]{Result: "some result"}
	if !resp.HasResult() {
		t.Errorf("Expected HasResult() to return true")
	}
}

func TestHasResultFalse(t *testing.T) {
	resp := jsonRPCResponse[string]{}
	if resp.HasResult() {
		t.Errorf("Expected HasResult() to return false")
	}
}

func TestHasErrorTrue(t *testing.T) {
	resp := jsonRPCResponse[string]{Error: &RPCError{Code: -1, Message: "error"}}
	if !resp.HasError() {
		t.Errorf("Expected HasError() to return true")
	}
}

func TestHasErrorFalse(t *testing.T) {
	resp := jsonRPCResponse[string]{}
	if resp.HasError() {
		t.Errorf("Expected HasError() to return false")
	}
}

func TestRPCErrorError(t *testing.T) {
	err := &RPCError{Code: -32601, Message: "Method not found", Data: nil}
	if err.Error() != "Method not found" {
		t.Errorf("Expected error message 'Method not found', got: %s", err.Error())
	}
}

func TestNewRPCError(t *testing.T) {
	data := map[string]string{"method": "unknown"}
	err := NewRPCError(-32601, "Method not found", data)

	if err.Code != -32601 {
		t.Errorf("Expected code -32601, got: %d", err.Code)
	}
	if err.Message != "Method not found" {
		t.Errorf("Expected message 'Method not found', got: %s", err.Message)
	}
	dataMap, ok := err.Data.(map[string]string)
	if !ok || dataMap["method"] != "unknown" {
		t.Errorf("Expected data map with method=unknown, got: %v", err.Data)
	}
}

func TestNewResponse(t *testing.T) {
	response := NewResponse("123", "success result")

	id, ok := response.GetID().(string)
	if !ok || id != "123" {
		t.Errorf("Expected ID '123', got: %v", response.GetID())
	}

	result, ok := response.GetResult().(string)
	if !ok || result != "success result" {
		t.Errorf("Expected result 'success result', got: %v", response.GetResult())
	}

	if !response.HasResult() {
		t.Error("Expected HasResult() to return true")
	}

	if response.HasError() {
		t.Error("Expected HasError() to return false")
	}
}

func TestUnmarshalJSONResponseValidationError(t *testing.T) {
	jsonData := []byte(`{"jsonrpc":"2.0","id":"123","result":"success","error":{"code":-32601,"message":"Method not found"}}`)
	var resp jsonRPCResponse[string]
	err := resp.UnmarshalJSON(jsonData)

	if err == nil {
		t.Error("Expected validation error for having both result and error")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("Expected ValidationError, got: %T", err)
	} else if validationErr.Error() != "response MUST NOT contain both result and error" {
		t.Errorf("Expected validation error message about both result and error, got: %s", validationErr.Error())
	}
}

func TestGetIDAndSetIDForResponse(t *testing.T) {
	// Создаем пустой ответ
	resp := &jsonRPCResponse[string]{}

	// Проверяем установку и получение ID
	testID := "test-id-123"
	err := resp.SetID(testID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	gotID := resp.GetID()
	if gotID != testID {
		t.Errorf("Expected ID %q, got %q", testID, gotID)
	}

	// Проверяем ошибку при установке неправильного типа ID
	err = resp.SetID(42) // подаем int вместо string
	if err == nil {
		t.Error("Expected error when setting wrong ID type, got nil")
	}
}

func TestGetResultAndSetResult(t *testing.T) {
	// Создаем пустой ответ
	resp := &jsonRPCResponse[string]{}

	// Проверяем установку и получение результата
	testResult := map[string]interface{}{"key": "value"}
	err := resp.SetResult(testResult)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	gotResult := resp.GetResult()
	resultMap, ok := gotResult.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be map[string]interface{}, got %T", gotResult)
	}

	if resultMap["key"] != "value" {
		t.Errorf("Expected result[\"key\"] == \"value\", got %v", resultMap["key"])
	}
}

func TestGetErrorAndSetError(t *testing.T) {
	// Создаем пустой ответ
	resp := &jsonRPCResponse[string]{}

	// Проверяем установку и получение ошибки
	testError := &RPCError{Code: -32700, Message: "Parse error"}
	resp.SetError(testError)

	gotError := resp.GetError()
	if gotError == nil {
		t.Fatal("Expected error to be non-nil")
	}

	if gotError.Code != -32700 || gotError.Message != "Parse error" {
		t.Errorf("Expected error {Code: -32700, Message: \"Parse error\"}, got %+v", gotError)
	}
}

func TestNewResponseWithStringID(t *testing.T) {
	// Создаем новый ответ с успешным результатом
	result := "success"
	resp := NewResponse("req-123", result)

	// Проверяем ID
	id, ok := resp.GetID().(string)
	if !ok || id != "req-123" {
		t.Errorf("Expected ID \"req-123\", got %v (%T)", resp.GetID(), resp.GetID())
	}

	// Проверяем результат
	gotResult := resp.GetResult()
	strResult, ok := gotResult.(string)
	if !ok || strResult != "success" {
		t.Errorf("Expected result \"success\", got %v (%T)", gotResult, gotResult)
	}

	// Проверяем, что ошибка отсутствует
	if resp.HasError() {
		t.Error("Expected HasError() to be false")
	}

	// Проверяем, что результат присутствует
	if !resp.HasResult() {
		t.Error("Expected HasResult() to be true")
	}
}

func TestNewResponseWithIntID(t *testing.T) {
	// Создаем новый ответ с int ID и успешным результатом
	result := map[string]interface{}{"status": "ok"}
	resp := NewResponse(123, result)

	// Проверяем ID
	id, ok := resp.GetID().(int)
	if !ok || id != 123 {
		t.Errorf("Expected ID 123, got %v (%T)", resp.GetID(), resp.GetID())
	}

	// Проверяем результат
	gotResult := resp.GetResult()
	mapResult, ok := gotResult.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be map[string]interface{}, got %T", gotResult)
	}

	if mapResult["status"] != "ok" {
		t.Errorf("Expected result[\"status\"] == \"ok\", got %v", mapResult["status"])
	}
}

// Test for creating a new error response
func TestNewErrorResponse(t *testing.T) {
	// Create a function that mirrors NewResponse but creates an error response
	newErrorResponse := func(id string, err *RPCError) Response {
		resp := &jsonRPCResponse[string]{
			JSONRPC: JSONRPCVersion,
			ID:      ID[string]{Value: id},
			Error:   err,
		}
		return resp
	}

	errCode := -32601
	errMessage := "Method not found"
	errData := map[string]string{"method": "unknown"}
	rpcErr := NewRPCError(errCode, errMessage, errData)

	response := newErrorResponse("123", rpcErr)

	// Verify ID
	id, ok := response.GetID().(string)
	if !ok || id != "123" {
		t.Errorf("Expected ID '123', got: %v", response.GetID())
	}

	// Verify error
	respError := response.GetError()
	if respError == nil {
		t.Fatal("Expected error to be non-nil")
	}
	if respError.Code != errCode {
		t.Errorf("Expected code %d, got: %d", errCode, respError.Code)
	}
	if respError.Message != errMessage {
		t.Errorf("Expected message '%s', got: %s", errMessage, respError.Message)
	}

	// Verify HasResult and HasError
	if response.HasResult() {
		t.Error("Expected HasResult() to return false")
	}
	if !response.HasError() {
		t.Error("Expected HasError() to return true")
	}
}

// Test JSON marshaling of response objects
func TestResponseJSONMarshaling(t *testing.T) {
	// Create a response with a result
	resp := &jsonRPCResponse[string]{
		JSONRPC: JSONRPCVersion,
		ID:      ID[string]{Value: "123"},
		Result:  "success",
	}

	// Marshal to JSON
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Check the JSON structure
	expected := `{"jsonrpc":"2.0","id":"123","result":"success"}`
	if string(data) != expected {
		t.Errorf("Expected JSON: %s, got: %s", expected, string(data))
	}

	// Now test with an error
	rpcErr := NewRPCError(-32700, "Parse error", nil)
	resp.Result = nil
	resp.Error = rpcErr

	data, err = json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}

	// Check the JSON structure for error response
	expected = `{"jsonrpc":"2.0","id":"123","error":{"code":-32700,"message":"Parse error"}}`
	if string(data) != expected {
		t.Errorf("Expected error JSON: %s, got: %s", expected, string(data))
	}
}

// Test handling invalid types for SetID
func TestSetIDWithInvalidTypes(t *testing.T) {
	resp := &jsonRPCResponse[string]{}

	// Try with float64 (should fail for string type)
	err := resp.SetID(42.5)
	if err == nil {
		t.Error("Expected error when setting ID with float64, got nil")
	}

	// Test with struct (should fail)
	type customStruct struct{}
	err = resp.SetID(customStruct{})
	if err == nil {
		t.Error("Expected error when setting ID with struct, got nil")
	}
}

// Test interactions between Result and Error fields
func TestResultErrorInteractions(t *testing.T) {
	resp := &jsonRPCResponse[string]{}

	// Test setting result then error (should work, validation happens separately)
	resp.SetResult("success")
	if !resp.HasResult() {
		t.Error("Expected HasResult() to be true after setting result")
	}

	rpcErr := NewRPCError(-32700, "Parse error", nil)
	resp.SetError(rpcErr)
	if !resp.HasError() {
		t.Error("Expected HasError() to be true after setting error")
	}

	// Test setting error to nil
	resp.SetError(nil)
	if resp.HasError() {
		t.Error("Expected HasError() to be false after setting error to nil")
	}

	// Test setting result to nil
	resp.SetResult(nil)
	if resp.HasResult() {
		t.Error("Expected HasResult() to be false after setting result to nil")
	}
}

// Test complex result types
func TestComplexResultTypes(t *testing.T) {
	resp := &jsonRPCResponse[string]{}

	// Test with map
	mapResult := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"nested": map[string]bool{
			"flag": true,
		},
	}

	err := resp.SetResult(mapResult)
	if err != nil {
		t.Errorf("Expected no error setting map result, got: %v", err)
	}

	result := resp.GetResult()
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be map[string]interface{}, got %T", result)
	}

	if resultMap["key1"] != "value1" || resultMap["key2"] != 42 {
		t.Errorf("Expected result map to match original, got: %v", resultMap)
	}

	// Test with slice
	sliceResult := []interface{}{"item1", 2, true}
	err = resp.SetResult(sliceResult)
	if err != nil {
		t.Errorf("Expected no error setting slice result, got: %v", err)
	}

	result = resp.GetResult()
	resultSlice, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected result to be []interface{}, got %T", result)
	}

	if len(resultSlice) != 3 || resultSlice[0] != "item1" {
		t.Errorf("Expected result slice to match original, got: %v", resultSlice)
	}
}

// Test RPCError creation and string conversion distinctly
func TestRPCErrorCreationAndConversion(t *testing.T) {
	customMessage := "Custom error message"
	customData := map[string]int{"code": 123}
	customCode := -12345

	err := NewRPCError(customCode, customMessage, customData)

	if err.Code != customCode {
		t.Errorf("Expected code %d, got: %d", customCode, err.Code)
	}

	if err.Message != customMessage {
		t.Errorf("Expected message '%s', got: '%s'", customMessage, err.Message)
	}

	if err.Error() != customMessage {
		t.Errorf("Error() should return message. Expected '%s', got: '%s'",
			customMessage, err.Error())
	}

	// Check that data was correctly stored
	dataMap, ok := err.Data.(map[string]int)
	if !ok {
		t.Fatalf("Expected data to be map[string]int, got %T", err.Data)
	}

	if dataMap["code"] != 123 {
		t.Errorf("Expected data[\"code\"] == 123, got %v", dataMap["code"])
	}
}
