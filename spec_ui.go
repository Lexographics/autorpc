package autorpc

import (
	"encoding/json"
	"net/http"
)

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
