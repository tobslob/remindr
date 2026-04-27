package main

import (
	"errors"
	"net/http"

	"github.com/tobslob/todoApp/cmd/utils"
	"github.com/tobslob/todoApp/internal/store"
)

func (app *application) AttachTagToTaskHandler(w http.ResponseWriter, r *http.Request) {

	taskID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	tagID, err := utils.GetIDParam(r, "tag_id")
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	taskTag := &store.TaskTag{
		TaskID: taskID,
		TagID:  tagID,
	}

	ctx := r.Context()

	if err := app.store.TaskTags.AttachTagToTask(ctx, taskTag); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		if errors.Is(err, store.ErrConflict) {
			utils.ConflictErr(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusCreated, taskTag); err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
}

func (app *application) GetTagsByTaskIDsHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	taskIDs, err := utils.GetIDsInQuery(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	tagsByTaskID, err := app.store.TaskTags.GetTagsByTaskIDs(ctx, taskIDs, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusOK, tagsByTaskID); err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
}

func (app *application) DetachTagFromTaskHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	taskID, err := utils.GetIDParam(r, "task_id")
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	tagID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.TaskTags.DetachTagFromTask(ctx, taskID, tagID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) GetTasksByTagIDHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	tagID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	tasks, err := app.store.TaskTags.GetTasksByTagID(ctx, tagID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusOK, tasks); err != nil {
		utils.InternalServerError(w, r, err)
		return
	}
}
