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

type MathService struct {
}

func (s *MathService) Add(ctx context.Context, params BinaryOpParams) (float32, error) {
	return params.A + params.B, nil
}

func (s *MathService) Subtract(ctx context.Context, params BinaryOpParams) (float32, error) {
	return params.A * params.B, nil
}

func (s *MathService) Multiply(ctx context.Context, params BinaryOpParams) (float32, error) {
	return params.A * params.B, nil
}

func (s *MathService) Divide(ctx context.Context, params BinaryOpParams) (float32, error) {
	return params.A / params.B, nil
}

func (s *MathService) Sum(ctx context.Context, params []float32) (float32, error) {
	sum := float32(0)
	for _, param := range params {
		sum += param
	}
	return sum, nil
}

func (s *MathService) Factorial(ctx context.Context, num int) (int, error) {
	if num < 0 {
		return 0, errors.New("factorial is not defined for non-positive numbers")
	}
	if num == 0 {
		return 1, nil
	}
	result := 1
	for i := 2; i <= num; i++ {
		result *= i
	}
	return result, nil
}

func (s *MathService) Fibonacci(ctx context.Context, num int) (int, error) {
	if num < 0 {
		return 0, errors.New("fibonacci is not defined for negative numbers")
	}
	if num == 0 || num == 1 {
		return num, nil
	}
	prev, curr := 0, 1
	for i := 2; i <= num; i++ {
		prev, curr = curr, prev+curr
	}
	return curr, nil
}

func main() {
	server := autorpc.NewServer()

	mathService := &MathService{}
	autorpc.RegisterMethod(server, "math.add", mathService.Add)
	autorpc.RegisterMethod(server, "math.subtract", mathService.Subtract)
	autorpc.RegisterMethod(server, "math.multiply", mathService.Multiply)
	autorpc.RegisterMethod(server, "math.divide", mathService.Divide)
	autorpc.RegisterMethod(server, "math.sum", mathService.Sum)
	autorpc.RegisterMethod(server, "math.factorial", mathService.Factorial)
	autorpc.RegisterMethod(server, "math.fibonacci", mathService.Fibonacci)

	http.Handle("/rpc", autorpc.HTTPHandler(server))
	http.Handle("/spec", autorpc.SpecUIHandler(server))
	http.Handle("/spec.json", autorpc.SpecJSONHandler(server))

	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
