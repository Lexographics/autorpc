# autorpc

A type-safe, JSON-RPC 2.0 server library for Go. With minimal boilerplate, automatic request handling, validation, middlewares, and introspection.

## Features

- **Type-Safe**: Uses Go generics for compile-time type checking
- **Automatic Validation**: Built-in parameter validation using struct tags
- **Introspection UI**: Built-in web UI for exploring your API

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
	
	// Set up HTTP handler
	http.Handle("/rpc", autorpc.HTTPHandler(server))
	
	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
```

Call your RPC method:

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

### Method with Struct Parameters

```go
type AddParams struct {
	A float32 `json:"a" validate:"required"`
	B float32 `json:"b" validate:"required"`
}

func Add(ctx context.Context, params AddParams) (float32, error) {
	return params.A + params.B, nil
}

autorpc.RegisterMethod(server, "math.add", Add)
```

### Using Groups and Middleware

```go
// Global middleware
server.Use(ExampleMiddleware(""))

// Create group with prefix and middleware
mathGroup := server.Group("math.", ExampleMiddleware("math"))

// Register methods in group
autorpc.RegisterMethod(mathGroup, "add", Add)
autorpc.RegisterMethod(mathGroup, "multiply", Multiply)

// Register with method-specific middleware
autorpc.RegisterMethod(mathGroup, "divide", Divide, ExampleMiddleware("divide"))
```

### Middleware

```go
func ExampleMiddleware(text string) autorpc.Middleware {
	return func(ctx context.Context, req autorpc.RPCRequest, next autorpc.HandlerFunc) (autorpc.RPCResponse, error) {
		log.Println(text)
		return next(ctx, req)
	}
}
```

### HTTP Context Access

```go
func MyMiddleware(ctx context.Context, req autorpc.RPCRequest, next autorpc.HandlerFunc) (autorpc.RPCResponse, error) {
	httpReq := autorpc.HTTPRequestFromContext(ctx)
	if httpReq != nil {
		userAgent := httpReq.Header.Get("User-Agent")
		cookie, _ := httpReq.Cookie("session_id")
	}
	return next(ctx, req)
}
```

## API Reference

### Server

```go
server := autorpc.NewServer()
server.Use(middlewares ...Middleware) // Add global middleware
```

### RegisterMethod

```go
autorpc.RegisterMethod[P, R any](r Registerer, name string, fn func(context.Context, P) (R, error), middlewares ...Middleware)
```

- `r`: Can be `*Server` or `*Group`
- `name`: Method name (prefix added automatically for groups)
- `fn`: Handler function
- `middlewares`: Optional method-specific middleware

### Groups

```go
group := server.Group(prefix string, middlewares ...Middleware)
group.Use(middlewares ...Middleware) // Add middleware to group
```

### Validation

Use struct tags from `github.com/go-playground/validator/v10`:

```go
type Params struct {
	Name string `json:"name" validate:"required,min=3"`
	Age  int    `json:"age" validate:"required,min=18"`
}
```

### Custom Errors

Implement `RPCErrorProvider`:

```go
type CustomError struct {
	code    int
	message string
	data interface{}
}

func (e *CustomError) Error() string { return e.message }
func (e *CustomError) Code() int     { return e.code }
func (e *CustomError) Message() string { return e.message }
func (e *CustomError) Data() interface{} { return e.data }
```

## Method Signature

All methods must follow this signature:

```go
func MethodName(ctx context.Context, params ParamsType) (ResultType, error)
```

- First parameter: `context.Context`
- Second parameter: Any type (primitive, struct, slice, pointer, etc.)
- Returns: `(ResultType, error)`

## Middleware Execution Order

1. Global middleware (`server.Use(...)`)
2. Group middleware (`group.Use(...)`)
3. Method-specific middleware (`RegisterMethod(..., middleware...)`)
4. Handler

**Note**: Middleware is captured at registration time. Methods registered before middleware is added won't have that middleware.

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
