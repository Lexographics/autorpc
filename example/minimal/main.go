package main

import (
	"log"
	"net/http"

	"github.com/Lexographics/autorpc"
)

func Greet(name string) (string, error) {
	return "Hello, " + name + "!", nil
}

func main() {
	server := autorpc.NewServer()

	autorpc.RegisterMethod(server, "greet", Greet)

	http.Handle("/rpc", autorpc.HTTPHandler(server))
	http.Handle("/spec.json", autorpc.SpecJSONHandler(server))

	log.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
