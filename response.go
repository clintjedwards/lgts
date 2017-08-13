package main

import (
	"encoding/json"
	"net/http"
)

// sendErrorResponse formats and sends an error message to supplied writer in json format
func sendErrorResponse(w http.ResponseWriter, httpStatusCode int, errorMessage string, err error) {
	w.WriteHeader(httpStatusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Ok         bool   `json:"ok"`
		StatusText string `json:"status_text"`
		Message    string `json:"message"`
		Error      string `json:"error"`
	}{false, http.StatusText(httpStatusCode), errorMessage, err.Error()})
}
