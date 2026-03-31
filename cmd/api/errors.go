package main

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s path: %s\n error: %s", r.Method, r.URL.Path, err)
	errorJson(w, http.StatusInternalServerError, err.Error())
}

func (app *application) conflictErr(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("conflict error: %s path: %s\n error: %s", r.Method, r.URL.Path, err)
	errorJson(w, http.StatusConflict, err.Error())
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request: %s path: %s\n error: %s", r.Method, r.URL.Path, err)
	errorJson(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found request: %s path: %s\n error: %s", r.Method, r.URL.Path, err)
	errorJson(w, http.StatusNotFound, err.Error())
}
