package main

import (
	"encoding/json"
	"net/http"
)

func writeJson(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func readJson(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578 // 1mb

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}

func errorJson(w http.ResponseWriter, status int, message string) error {

	type jsonError struct {
		Error string `json:"error"`
	}

	return writeJson(w, status, &jsonError{Error: message})
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type jsonResponse struct {
		Data any `json:"data"`
	}

	return writeJson(w, status, &jsonResponse{Data: data})
}
