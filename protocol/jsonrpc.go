package protocol

// import (
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// )

// // JSON-RPC Version
// // const JSONRPCVersion = "2.0"

// // type IDType interface{}

// // https://www.jsonrpc.org/specification#request_object
// // type JSONRPCRequest struct {
// // 	// A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
// // 	JSONRPC string `json:"jsonrpc"`
// // 	// A String containing the name of the method to be invoked.
// // 	// Method names that begin with the word rpc followed by a period character (U+002E or ASCII 46) are reserved for rpc-internal methods
// // 	// and extensions and MUST NOT be used for anything else.
// // 	Method string `json:"method"`
// // 	// A Structured value that holds the parameter values to be used during the invocation of the method.
// // 	// This member MAY be omitted.
// // 	Params interface{} `json:"params,omitempty"` // optional parameters
// // 	// An identifier established by the Client that MUST contain a String, Number, or NULL value if included.
// // 	// If it is not included it is assumed to be a notification.
// // 	// The value SHOULD normally not be Null [1] and Numbers SHOULD NOT contain fractional parts [2]
// // 	// The Server MUST reply with the same value in the Response object if included.
// // 	// This member is used to correlate the context between the two objects.
// // 	ID IDType `json:"id"`
// // }

// // https://www.jsonrpc.org/specification#response_object
// // Either the result member or error member MUST be included, but both members MUST NOT be included.
// type JSONRPCResponse struct {
// 	// A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
// 	JSONRPC string `json:"jsonrpc"`
// 	// This member is REQUIRED on success.
// 	// This member MUST NOT exist if there was an error invoking the method.
// 	// The value of this member is determined by the method invoked on the Server.
// 	Result interface{} `json:"result,omitempty"`
// 	// This member is REQUIRED on error.
// 	// This member MUST NOT exist if there was no error triggered during invocation.
// 	// The value for this member MUST be an Object as defined in section 5.1.
// 	Error *JSONRPCError `json:"error,omitempty"`
// 	// This member is REQUIRED.
// 	// It MUST be the same as the value of the id member in the Request Object.
// 	// If there was an error in detecting the id in the Request object (e.g. Parse error/Invalid Request), it MUST be Null.
// 	ID IDType `json:"id"`
// }

// // https://www.jsonrpc.org/specification#error_object
// type JSONRPCError struct {
// 	// A Number that indicates the error type that occurred.
// 	// This MUST be an integer.
// 	Code int `json:"code"`
// 	// A String providing a short description of the error.
// 	// The message SHOULD be limited to a concise single sentence.
// 	Message string `json:"message"`
// 	// A Primitive or Structured value that contains additional information about the error.
// 	// This may be omitted.
// 	// The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.).
// 	Data interface{} `json:"data,omitempty"`
// }

// func (e *JSONRPCError) Error() string {
// 	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
// }

// // NewResponse creates a JSONRPCResponse with a result.
// func NewResponse(id IDType, result interface{}) *JSONRPCResponse {
// 	return &JSONRPCResponse{
// 		JSONRPC: JSONRPCVersion,
// 		ID:      id,
// 		Result:  result,
// 	}
// }

// // NewErrorResponse creates a JSONRPCResponse with an error.
// func NewErrorResponse(id IDType, code int, message string, data interface{}) *JSONRPCResponse {
// 	return &JSONRPCResponse{
// 		JSONRPC: JSONRPCVersion,
// 		ID:      id,
// 		Error:   &JSONRPCError{Code: code, Message: message, Data: data},
// 	}
// }

// func (r *JSONRPCResponse) Validate() error {
// 	if r.JSONRPC != JSONRPCVersion {
// 		return errors.New("invalid JSON-RPC version")
// 	}
// 	if r.Result != nil && r.Error != nil {
// 		return errors.New("response cannot have both result and error")
// 	}
// 	if r.Result == nil && r.Error == nil {
// 		return errors.New("response must have either result or error")
// 	}
// 	return nil
// }

// // func (r *JSONRPCRequest) Validate() error {
// // 	if r.JSONRPC != JSONRPCVersion {
// // 		return errors.New("invalid JSON-RPC version")
// // 	}
// // 	if r.Method == "" {
// // 		return errors.New("method is required")
// // 	}
// // 	return nil
// // }

// func (r *JSONRPCRequest) IsNotification() bool {
// 	return r.ID == nil
// }

// func ParseRequest(jsonData []byte) (*JSONRPCRequest, error) {
// 	var req JSONRPCRequest
// 	if err := json.Unmarshal(jsonData, &req); err != nil {
// 		return nil, &JSONRPCError{Code: ParseError, Message: "failed to parse JSON", Data: err.Error()}
// 	}
// 	if err := req.Validate(); err != nil {
// 		return nil, &JSONRPCError{Code: InvalidRequest, Message: err.Error()}
// 	}
// 	return &req, nil
// }

// func ParseResponse(jsonData []byte) (*JSONRPCResponse, error) {
// 	var resp JSONRPCResponse
// 	if err := json.Unmarshal(jsonData, &resp); err != nil {
// 		return nil, &JSONRPCError{Code: ParseError, Message: "failed to parse JSON", Data: err.Error()}
// 	}
// 	if err := resp.Validate(); err != nil {
// 		return nil, &JSONRPCError{Code: InvalidRequest, Message: err.Error()}
// 	}
// 	return &resp, nil
// }
