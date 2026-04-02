package main

import (
	"errors"
	"net/http"

	"github.com/tobslob/todoApp/cmd/utils"
	"github.com/tobslob/todoApp/internal/store"
)

type CreateItemPayload struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (app *application) CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	var payload CreateItemPayload

	if err := utils.ReadJson(w, r, &payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	item := &store.Item{
		UserID:      user.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Status:      store.Todo,
	}

	ctx := r.Context()

	if err := app.store.Items.Create(ctx, item); err != nil {
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusCreated, item); err != nil {
		utils.InternalServerError(w, r, err)
	}
}
