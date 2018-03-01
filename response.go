package main

import (
	"encoding/json"
	"net/http"
)

//sendResponse formats and sends an error message to supplied writer in json format
func sendResponse(w http.ResponseWriter, httpStatusCode int, data interface{}, error bool) error {

	if httpStatusCode != 200 {
		w.WriteHeader(httpStatusCode)
	}

	if error {
		err := json.NewEncoder(w).Encode(struct {
			StatusText string      `json:"status_text"`
			Message    interface{} `json:"message"`
		}{http.StatusText(httpStatusCode), data})
		if err != nil {
			return err
		}
		return nil
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(data)

	return nil
}
