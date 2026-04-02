package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func WriteJson(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func ReadJson(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578 // 1mb

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(data); err != nil {
		if err == io.EOF {
			return errors.New("body must not be empty")
		}
		return err
		}
	return nil
}

func ErrorJson(w http.ResponseWriter, status int, message string) error {

	type jsonError struct {
		Error string `json:"error"`
	}

	return WriteJson(w, status, &jsonError{Error: message})
}

func JsonResponse(w http.ResponseWriter, status int, data any) error {
	type jsonResponse struct {
		Data any `json:"data"`
	}

	return WriteJson(w, status, &jsonResponse{Data: data})
}
