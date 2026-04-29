package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/tobslob/remindr/cmd/utils"
	"github.com/tobslob/remindr/internal/store"
)

type CreateUserPayload struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100,alphanum"`
}

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (app *application) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserPayload

	if err := utils.ReadJson(w, r, &payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}

	user := &store.User{
		Password: hashedPassword,
		Email:    payload.Email,
		Username: payload.Username,
	}

	ctx := r.Context()

	if err := app.store.Users.Create(ctx, user); err != nil {
		if errors.Is(err, store.ErrConflict) {
			utils.ConflictErr(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusCreated, user); err != nil {
		utils.InternalServerError(w, r, err)
	}
}

func (app *application) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var login LoginPayload

	if err := utils.ReadJson(w, r, &login); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if err := utils.Validate.Struct(login); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	user, err := app.store.Users.GetByEmail(ctx, login.Email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.UnauthorizedError(w, r, errors.New("invalid email or password"))
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if ok := utils.CheckPassword(login.Password, user.Password); !ok {
		utils.UnauthorizedError(w, r, errors.New("invalid email or password"))
		return
	}

	expirationTime, err := time.ParseDuration("24h")
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}

	token, _, err := app.tokenMaker.CreateToken(user.ID, expirationTime)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusOK, map[string]any{"token": token}); err != nil {
		utils.InternalServerError(w, r, err)
	}
}
