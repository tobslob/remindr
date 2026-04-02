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

func (app *application) GetItemsHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	ctx := r.Context()

	items, err := app.store.Items.GetItems(ctx, user.ID)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusOK, items); err != nil {
		utils.InternalServerError(w, r, err)
	}
}

func (app *application) GetItemByIDHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	itemID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	item, err := app.store.Items.GetByID(ctx, itemID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusOK, item); err != nil {
		utils.InternalServerError(w, r, err)
	}
}

func (app *application) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	itemID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.Items.DeleteByID(ctx, itemID, user.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) UpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	itemID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
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

	ctx := r.Context()

	existingItem, err := app.store.Items.GetByID(ctx, itemID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	item := &store.Item{
		ID:          itemID,
		UserID:      user.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Status:      existingItem.Status,
	}

	if err := app.store.Items.UpdateByID(ctx, user.ID, item); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	updatedItem, err := app.store.Items.GetByID(ctx, itemID, user.ID)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusOK, updatedItem); err != nil {
		utils.InternalServerError(w, r, err)
	}
}

func (app *application) DeleteByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	ctx := r.Context()

	if err := app.store.Items.DeleteAllByUserID(ctx, user.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) DeleteByIDsHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	ids, err := utils.GetIDsInQuery(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.Items.DeleteByIDs(ctx, ids, user.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
