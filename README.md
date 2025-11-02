# autorpc

A type-safe, JSON-RPC 2.0 server library for Go. With minimal boilerplate, automatic request handling, validation, and introspection.

## Features

- **Type-Safe**: Use Go generics for compile-time type checking
- **Automatic Validation**: Built-in parameter validation using struct tags
- **Custom Errors**: Define custom error codes and messages
- **Introspection UI (wip)**: Built-in web UI for exploring your API
- **Method Specs**: JSON API for programmatic access to method information
- **Zero Boilerplate**: Register methods with simple function signatures

## Installation

```bash
go get github.com/Lexographics/autorpc
```

## Quick Start

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Lexographics/autorpc"
)

func Greet(ctx context.Context, name string) (string, error) {
	return "Hello, " + name + "!", nil
}

func main() {
	server := autorpc.NewServer()

	// Register a method
	autorpc.RegisterMethod(server, "greet", Greet)

	// Set up HTTP handlers
	http.Handle("/rpc", autorpc.HTTPHandler(server))
	http.Handle("/spec", autorpc.SpecUIHandler(server))
	http.Handle("/spec.json", autorpc.SpecJSONHandler(server))

	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
```

Now you can call your RPC method:

```bash
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"greet","params":"John","id":1}'
```

Response:
```json
{
  "jsonrpc": "2.0",
  "result": "Hello, John!",
  "id": 1
}
```

## Examples

### Basic Method with Primitive Parameters

```go
func Add(ctx context.Context, params []float32) (float32, error) {
	if len(params) != 2 {
		return 0, errors.New("expected exactly 2 numbers")
	}
	return params[0] + params[1], nil
}

autorpc.RegisterMethod(server, "math.add", Add)
```

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "math.add",
  "params": [5.5, 3.2],
  "id": 1
}
```

### Method with Struct Parameters

```go
type AddParams struct {
	A float32 `json:"a"`
	B float32 `json:"b"`
}

func Add(ctx context.Context, params AddParams) (float32, error) {
	return params.A + params.B, nil
}

autorpc.RegisterMethod(server, "math.add", Add)
```

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "math.add",
  "params": {"a": 5.5, "b": 3.2},
  "id": 1
}
```

### Parameter Validation

Use struct tags from `github.com/go-playground/validator/v10` for automatic validation:

```go
type ConcatParams struct {
	A string `json:"a" validate:"required"`
	B string `json:"b" validate:"required,min=1"`
}

func Concat(ctx context.Context, params ConcatParams) (string, error) {
	return params.A + params.B, nil
}

autorpc.RegisterMethod(server, "string.concat", Concat)
```

If validation fails, the server automatically returns a proper JSON-RPC error response:

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": [
      {
        "field": "A",
        "tag": "required",
      }
    ]
  },
  "id": 1
}
```

### Custom Validation Error Handler

You can customize how validation errors are formatted:

```go
import "github.com/go-playground/validator/v10"

server.SetValidateErrorHandler(func(errs *validator.ValidationErrors) *autorpc.RPCError {
	details := make([]string, 0, len(*errs))
	for _, err := range *errs {
		details = append(details, fmt.Sprintf("%s: %s", err.Field(), err.Tag()))
	}
	return &autorpc.RPCError{
		Code:    autorpc.CodeInvalidParams,
		Message: "Validation failed",
		Data:    details,
	}
})
```

### Custom Error Types

Implement the `RPCErrorProvider` interface to return custom JSON-RPC error codes:

```go
type CustomError struct {
	code    int
	message string
	data    interface{}
}

func (e *CustomError) Error() string {
	return e.message
}

func (e *CustomError) Code() int {
	return e.code
}

func (e *CustomError) Message() string {
	return e.message
}

func (e *CustomError) Data() interface{} {
	return e.data
}

func CustomMethod(ctx context.Context, params string) (any, error) {
	return nil, &CustomError{
		code:    -32000, // Custom error code
		message: "This is a custom error",
		data:    map[string]string{"param": params},
	}
}

autorpc.RegisterMethod(server, "debug.custom-error", CustomMethod)
```

### Using Methods from Structs

You can register methods from struct instances:

```go
type MathService struct{}

func (s *MathService) Add(ctx context.Context, params AddParams) (float32, error) {
	return params.A + params.B, nil
}

func (s *MathService) Multiply(ctx context.Context, params AddParams) (float32, error) {
	return params.A * params.B, nil
}

func main() {
	server := autorpc.NewServer()
	mathService := &MathService{}

	autorpc.RegisterMethod(server, "math.add", mathService.Add)
	autorpc.RegisterMethod(server, "math.multiply", mathService.Multiply)

	// ... setup HTTP handlers
}
```

### Batch Requests

Send multiple requests in a single HTTP call:

```json
[
  {"jsonrpc":"2.0","method":"math.add","params":{"a":1,"b":2},"id":1},
  {"jsonrpc":"2.0","method":"math.multiply","params":{"a":3,"b":4},"id":2}
]
```

The server processes batch requests in parallel and returns an array of responses:

```json
[
  {"jsonrpc":"2.0","result":3,"id":1},
  {"jsonrpc":"2.0","result":12,"id":2}
]
```

### Notifications

Send a notification (fire-and-forget request) by omitting the `id` field:

```json
{
  "jsonrpc": "2.0",
  "method": "notify",
  "params": {"message": "Hello"}
}
```

The server will return HTTP 204 No Content (no response body).

## API Reference

### Server

#### `NewServer() *Server`

Creates a new JSON-RPC server instance.

```go
server := autorpc.NewServer()
```

#### `RegisterMethod[P, R any](s *Server, name string, fn func(context.Context, P) (R, error))`

Registers an RPC method with the server. The function must have exactly two parameters (context.Context and params) and return exactly two values (result, error).

- `P`: The parameter type (can be a primitive, struct, slice, etc.)
- `R`: The result type
- `name`: The method name (e.g., "math.add")
- `fn`: The function to call when this method is invoked

```go
autorpc.RegisterMethod(server, "greet", Greet)
```

#### `SetValidateErrorHandler(handler ValidateErrorHandler)`

Sets a custom handler for validation errors. If not set, uses the default handler.

```go
server.SetValidateErrorHandler(func(errs *validator.ValidationErrors) *autorpc.RPCError {
	// Return a custom error
	return &autorpc.RPCError{...}
})
```

#### `GetMethodSpecs() []MethodInfo`

Returns information about all registered methods. Useful for API documentation or introspection.

```go
specs := server.GetMethodSpecs()
for _, spec := range specs {
	fmt.Printf("Method: %s\n", spec.Name)
	fmt.Printf("Params: %+v\n", spec.Params)
	fmt.Printf("Result: %+v\n", spec.Result)
}
```

### HTTP Handlers

#### `HTTPHandler(server *Server) http.Handler`

Returns an HTTP handler that processes JSON-RPC requests. Handles both single requests and batch requests.

```go
http.Handle("/rpc", autorpc.HTTPHandler(server))
```

#### `SpecJSONHandler(server *Server) http.Handler`

Returns an HTTP handler that serves method specifications as JSON.

```go
http.Handle("/spec.json", autorpc.SpecJSONHandler(server))
```

### Error Handling

#### Standard JSON-RPC Error Codes

```go
const (
	CodeParseError     = -32700  // Invalid JSON was received
	CodeInvalidRequest = -32600  // The JSON sent is not a valid Request object
	CodeMethodNotFound = -32601  // The method does not exist
	CodeInvalidParams  = -32602  // Invalid method parameter(s)
	CodeInternalError  = -32603  // Internal JSON-RPC error
)
```

#### `RPCErrorProvider` Interface

Implement this interface to return custom error codes:

```go
type RPCErrorProvider interface {
	Code() int
	Message() string
	Data() interface{}
}
```

## Method Signature Requirements

All registered methods must follow this signature:

```go
func MethodName(ctx context.Context, params ParamsType) (ResultType, error)
```

- **Context**: First parameter must be `context.Context`
- **Params**: Any of the supported types
- **Returns**: Exactly two values - the result (any type) and an error

**Supported Parameter Types:**
- Primitives: `string`, `int`, `float32`, `float64`, `bool`, etc.
- Slices/Arrays: `[]string`, `[]int`, etc.
- Structs: Any struct type with JSON tags
- Pointers: `*MyType` (unmarshalled automatically)

**Example Signatures:**

```go
// Primitive parameter
func Greet(ctx context.Context, name string) (string, error)

// Slice parameter
func Sum(ctx context.Context, numbers []float32) (float32, error)

// Struct parameter
func Add(ctx context.Context, params AddParams) (float32, error)

// Pointer parameter
func Process(ctx context.Context, data *MyData) (*MyResult, error)
```

## Complete Example

Here's a complete example with multiple features:

```go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/Lexographics/autorpc"
)

type BinaryOpParams struct {
	A float32 `json:"a" validate:"required"`
	B float32 `json:"b" validate:"required"`
}

type MathService struct{}

func (s *MathService) Add(ctx context.Context, params BinaryOpParams) (float32, error) {
	return params.A + params.B, nil
}

func (s *MathService) Divide(ctx context.Context, params BinaryOpParams) (float32, error) {
	if params.B == 0 {
		return 0, errors.New("division by zero")
	}
	return params.A / params.B, nil
}

func (s *MathService) Sum(ctx context.Context, numbers []float32) (float32, error) {
	sum := float32(0)
	for _, n := range numbers {
		sum += n
	}
	return sum, nil
}

func main() {
	server := autorpc.NewServer()

	mathService := &MathService{}
	autorpc.RegisterMethod(server, "math.add", mathService.Add)
	autorpc.RegisterMethod(server, "math.divide", mathService.Divide)
	autorpc.RegisterMethod(server, "math.sum", mathService.Sum)

	http.Handle("/rpc", autorpc.HTTPHandler(server))
	http.Handle("/spec", autorpc.SpecUIHandler(server))
	http.Handle("/spec.json", autorpc.SpecJSONHandler(server))

	log.Println("Server started on port 8080")
	log.Println("RPC endpoint: http://localhost:8080/rpc")
	log.Println("Spec UI: http://localhost:8080/spec")
	log.Println("Spec JSON: http://localhost:8080/spec.json")
	http.ListenAndServe(":8080", nil)
}
```

## More Examples

See the `example/` directory for additional examples:

- **minimal**: Basic "Hello World" example
- **main**: Examples with validation and custom errors
- **math**: A complete math service with multiple operations

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

