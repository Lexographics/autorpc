package autorpc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

func HTTPHandler(server *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			resp := newErrorResponse(nil, CodeParseError, "Failed to read body")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}
		defer r.Body.Close()

		trimmedBody := bytes.TrimSpace(body)

		if len(trimmedBody) == 0 {
			resp := newErrorResponse(nil, CodeInvalidRequest, "Empty request body")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}

		// starts with '[' -> batch
		// starts with '{' -> single
		if trimmedBody[0] == '[' {
			handleBatchHTTP(server, w, trimmedBody)
		} else if trimmedBody[0] == '{' {
			handleSingleHTTP(server, w, trimmedBody)
		} else {
			resp := newErrorResponse(nil, CodeParseError, "Invalid JSON")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
		}
	})
}

func handleSingleHTTP(server *Server, w http.ResponseWriter, body []byte) {
	var req RPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		resp := newErrorResponse(nil, CodeParseError, "Failed to parse JSON request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := server.processRequest(req)

	// If req.ID is nil, it's a Notification.
	// 4.1 Notification: "The Server MUST NOT reply to a Notification"
	if req.ID == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleBatchHTTP(server *Server, w http.ResponseWriter, body []byte) {
	var reqs []RPCRequest
	if err := json.Unmarshal(body, &reqs); err != nil {
		resp := newErrorResponse(nil, CodeParseError, "Failed to parse JSON batch")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if len(reqs) == 0 {
		resp := newErrorResponse(nil, CodeInvalidRequest, "Empty batch")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	responses := make([]RPCResponse, 0, len(reqs))
	var responsesMu sync.Mutex
	var wg sync.WaitGroup

	for _, req := range reqs {
		wg.Add(1)
		go func(r RPCRequest) {
			defer wg.Done()

			resp := server.processRequest(r)

			// 4.1 Notification: "The Server MUST NOT reply to a Notification"
			if r.ID != nil {
				responsesMu.Lock()
				responses = append(responses, resp)
				responsesMu.Unlock()
			}
		}(req)
	}

	wg.Wait()

	// If the batch only contains notifications, we must not return an empty array
	if len(responses) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}
