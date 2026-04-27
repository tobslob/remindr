package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tobslob/todoApp/cmd/tokens"
	"github.com/tobslob/todoApp/internal/store"
)

type application struct {
	config config
	store  *store.Storage
	tokenMaker tokens.Maker
}

type config struct{
	addr 					 string
	db 						 string
	dbMaxOpenConns int
	dbMaxIdleConns int
	dbMaxIdleTime  string
	TokenSecretKey string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthz", app.healthCheckHandler)
		r.Route("/users", func(r chi.Router) {
			r.Post("/register", app.CreateUserHandler)
			r.Post("/login", app.LoginUserHandler)
		})

		r.Route("/tasks", func(r chi.Router) {
			r.Use(app.AuthMiddleware)
			r.Post("/", app.CreateTaskHandler)
			r.Post("/{id}/tags/{tag_id}", app.AttachTagToTaskHandler)
			r.Get("/tags", app.GetTagsByTaskIDsHandler)
			r.Get("/{id}", app.GetTaskByIDHandler)
			r.Get("/", app.GetTasksHandler)
			r.Patch("/{id}", app.UpdateTaskHandler)
			r.Delete("/{id}", app.DeleteTaskHandler)
			r.Delete("/", app.DeleteByUserIDHandler)
			r.Delete("/bulk", app.DeleteByIDsHandler)
		})

		r.Route("/tags", func(r chi.Router) {
			r.Use(app.AuthMiddleware)
			r.Post("/", app.CreateTagHandler)
			r.Get("/", app.GetTagsHandler)
			r.Get("/{id}", app.GetTagHandler)
			r.Get("/{id}/tasks", app.GetTasksByTagIDHandler)
			r.Patch("/{id}", app.UpdateTagHandler)
			r.Delete("/{id}", app.DeleteTagHandler)
			r.Delete("/{task_id}/{id}", app.DetachTagFromTaskHandler)
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server has started on port %s", app.config.addr)

	return srv.ListenAndServe()
}
