package autorpc

import (
	_ "embed"
	"encoding/json"
	"net/http"
)

//go:embed spec_ui.html
var specUIString string

func SpecUIHandler(server *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(specUIString))
	})
}

func SpecJSONHandler(server *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		specs := server.GetMethodSpecs()
		json.NewEncoder(w).Encode(specs)
	})
}
