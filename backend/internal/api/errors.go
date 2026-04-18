package api

import (
	"encoding/json"
	"net/http"
)

// APIError represents a standardized JSON error response.
type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// WriteError writes a structured JSON error to the response.
func WriteError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIError{
		Message: message,
		Code:    code,
	})
}
