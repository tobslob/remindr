package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/tobslob/todoApp/cmd/utils"
	"github.com/tobslob/todoApp/internal/store"
)

type CreateTaskPayload struct {
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Priority    *string `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueAt	   		time.Time `json:"due_at" validate:"datetime=2006-01-02T15:04:05Z07:00,required"`
}

func (app *application) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	var payload CreateTaskPayload

	if err := utils.ReadJson(w, r, &payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if payload.Priority == nil {
		p := "medium"
		payload.Priority = &p
	}

	item := &store.Task{
		UserID:      user.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Status:      store.Todo,
		Priority:    payload.Priority,
		DueAt:       payload.DueAt,
	}

	ctx := r.Context()

	if err := app.store.Tasks.Create(ctx, item); err != nil {
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusCreated, item); err != nil {
		utils.InternalServerError(w, r, err)
	}
}

func (app *application) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	taskListQuery, err := utils.GetTaskListQuery(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	tasks, err := app.store.Tasks.GetTasks(ctx, user.ID, store.TaskFilter{
		LastID:          taskListQuery.LastID,
		Limit:           taskListQuery.Limit,
		Search:          taskListQuery.Search,
		Status:          taskListQuery.Status,
		Priority:        taskListQuery.Priority,
		CreatedFrom:     taskListQuery.CreatedFrom,
		CreatedBefore:   taskListQuery.CreatedTo,
		DueFrom:         taskListQuery.DueFrom,
		DueBefore:       taskListQuery.DueTo,
		CompletedFrom:   taskListQuery.CompletedFrom,
		CompletedBefore: taskListQuery.CompletedTo,
	})
	if err != nil {
		if errors.Is(err, store.ErrInvalidCursor) {
			utils.BadRequestError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusOK, tasks); err != nil {
		utils.InternalServerError(w, r, err)
	}
}

func (app *application) GetTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
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

	item, err := app.store.Tasks.GetByID(ctx, itemID, user.ID)
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

func (app *application) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := app.store.Tasks.DeleteByID(ctx, itemID, user.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	taskID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	var payload CreateTaskPayload

	if err := utils.ReadJson(w, r, &payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	existingItem, err := app.store.Tasks.GetByID(ctx, taskID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if payload.Priority == nil {
		payload.Priority = existingItem.Priority
	}

	item := &store.Task{
		ID:          taskID,
		UserID:      user.ID,
		Title:       payload.Title,
		Description: payload.Description,
		Status:      existingItem.Status,
		Priority:    payload.Priority,
	}

	if err := app.store.Tasks.UpdateByID(ctx, user.ID, item); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	updatedItem, err := app.store.Tasks.GetByID(ctx, taskID, user.ID)
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

	if err := app.store.Tasks.DeleteAllByUserID(ctx, user.ID); err != nil {
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

	if err := app.store.Tasks.DeleteByIDs(ctx, ids, user.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
