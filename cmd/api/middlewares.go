package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/tobslob/remindr/cmd/utils"
	"github.com/tobslob/remindr/internal/requestctx"
)

func (app *application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.UnauthorizedError(w, r, errors.New("authorization header is required"))
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		payload, err := app.tokenMaker.VerifyToken(tokenStr)
		if err != nil {
			utils.UnauthorizedError(w, r, errors.New("invalid token"))
			return
		}

		user, err := app.store.Users.GetByID(r.Context(), payload.UserID)
		if err != nil {
			utils.UnauthorizedError(w, r, errors.New("user not found"))
			return
		}

		ctx := requestctx.WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
