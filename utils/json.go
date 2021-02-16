package utils

import (
	"encoding/json"
	"net/http"
)

// WriteErrorResponse is a helper function that returns JSON response for errors
func WriteErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	encodeError := json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
	if encodeError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Something bad happened, contact the system admin"}`))
	}
}
