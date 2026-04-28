package main

import (
	"errors"
	"net/http"

	"github.com/tobslob/todoApp/cmd/utils"
	"github.com/tobslob/todoApp/internal/store"
)

func (app *application) AttachTagToTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := utils.GetUserFromContext(ctx)
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	taskID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	task, err := app.store.Tasks.GetByID(ctx, taskID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if task.UserID != user.ID {
		utils.UnauthorizedError(w, r, errors.New("user does not have access to this task"))
		return
	}

	tagID, err := utils.GetIDParam(r, "tag_id")
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	tag, err := app.store.Tags.GetByID(ctx, tagID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if tag.UserID != user.ID {
		utils.UnauthorizedError(w, r, errors.New("user does not have access to this tag"))
		return
	}

	taskTag := &store.TaskTag{
		TaskID: taskID,
		TagID:  tagID,
	}

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
	ctx := r.Context()
	user := utils.GetUserFromContext(ctx)
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	taskIDs, err := utils.GetIDsInQuery(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

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
	ctx := r.Context()
	user := utils.GetUserFromContext(ctx)
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	taskID, err := utils.GetIDParam(r, "task_id")
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	task, err := app.store.Tasks.GetByID(ctx, taskID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if task.UserID != user.ID {
		utils.UnauthorizedError(w, r, errors.New("user does not have access to this task"))
		return
	}

	tagID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	tag, err := app.store.Tags.GetByID(ctx, tagID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if tag.UserID != user.ID {
		utils.UnauthorizedError(w, r, errors.New("user does not have access to this tag"))
		return
	}

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
	ctx := r.Context()
	user := utils.GetUserFromContext(ctx)
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	tagID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

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
