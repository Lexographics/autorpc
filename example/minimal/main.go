package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Lexographics/autorpc"
)

type GreetParams struct {
	Name string `json:"name" validate:"required"`
}

func Greet(ctx context.Context, params GreetParams) (string, error) {
	return "Hello, " + params.Name + "!", nil
}

func main() {
	server := autorpc.NewServer()

	autorpc.RegisterMethod(server, "greet", Greet)

	http.Handle("/rpc", autorpc.HTTPHandler(server))
	http.Handle("/spec", autorpc.SpecUIHandler(server))
	http.Handle("/spec.json", autorpc.SpecJSONHandler(server))

	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
