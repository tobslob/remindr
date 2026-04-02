package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/tobslob/todoApp/cmd/utils"
)

type userKey string

const USERCTX userKey = "user"

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

		ctx := context.WithValue(r.Context(), USERCTX, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
