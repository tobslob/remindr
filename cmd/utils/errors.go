package utils

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s path: %s\n error: %s", r.Method, r.URL.Path, err)
	ErrorJson(w, http.StatusInternalServerError, err.Error())
}

func ConflictErr(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("conflict error: %s path: %s\n error: %s", r.Method, r.URL.Path, err)
	ErrorJson(w, http.StatusConflict, err.Error())
}

func BadRequestError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request: %s path: %s\n error: %s", r.Method, r.URL.Path, err)
	ErrorJson(w, http.StatusBadRequest, err.Error())
}

func NotFoundError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found request: %s path: %s\n error: %s", r.Method, r.URL.Path, err)
	ErrorJson(w, http.StatusNotFound, err.Error())
}

func UnauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Unauthorized request: %s path: %s\n error: %s", r.Method, r.URL.Path, err)
	ErrorJson(w, http.StatusUnauthorized, err.Error())
}
