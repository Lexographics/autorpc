package autorpc

import (
	"context"
	"encoding/json"
)

type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      json.RawMessage `json:"id"`
}

type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      json.RawMessage `json:"id"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

func newErrorResponse(id json.RawMessage, code int, message string) RPCResponse {
	return RPCResponse{
		JSONRPC: "2.0",
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
}

type FuncType[P, R any] func(ctx context.Context, params P) (R, error)

// UnmarshalKind is an interface that types can implement to specify their JSON unmarshaling kind.
// This is useful for types like types.Time or types.Duration that unmarshal from strings or numbers
// but are represented as structs in Go. If a type implements this interface, the returned kind
// will be used in the API specification instead of "struct".
type UnmarshalKind interface {
	UnmarshalKind() string
}
