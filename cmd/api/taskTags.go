package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tobslob/todoApp/cmd/utils"
	"github.com/tobslob/todoApp/internal/store"
)

type ReplaceTaskTagsPayload struct {
	TagIDs []string `json:"tag_ids"`
}

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

func normalizeAndDeduplicateUUIDs(rawIDs []string) ([]uuid.UUID, error) {
	normalizedIDs := make([]uuid.UUID, 0, len(rawIDs))
	seen := make(map[uuid.UUID]struct{}, len(rawIDs))

	for _, rawID := range rawIDs {
		id, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil {
			return nil, err
		}

		if _, exists := seen[id]; exists {
			continue
		}

		seen[id] = struct{}{}
		normalizedIDs = append(normalizedIDs, id)
	}

	return normalizedIDs, nil
}

func (app *application) ReplaceTaskTagsHandler(w http.ResponseWriter, r *http.Request) {
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

	var payload ReplaceTaskTagsPayload
	if err := utils.ReadJson(w, r, &payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if payload.TagIDs == nil {
		utils.BadRequestError(w, r, errors.New("tag_ids is required"))
		return
	}

	tagIDs, err := normalizeAndDeduplicateUUIDs(payload.TagIDs)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if err := app.store.TaskTags.ReplaceTaskTags(ctx, taskID, user.ID, tagIDs); err != nil {
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

	tagsByTaskID, err := app.store.TaskTags.GetTagsByTaskIDs(ctx, []uuid.UUID{taskID}, user.ID)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}

	updatedTags := tagsByTaskID[taskID]
	if updatedTags == nil {
		updatedTags = []*store.Tag{}
	}

	if err := utils.JsonResponse(w, http.StatusOK, updatedTags); err != nil {
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
