package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tobslob/remindr/cmd/utils"
	"github.com/tobslob/remindr/internal/reminder"
	"github.com/tobslob/remindr/internal/store"
)

type CreateReminderPayload struct {
	TaskID   uuid.UUID             `json:"task_id" validate:"required"`
	Type     reminder.ReminderType `json:"type" validate:"required,oneof=before_due due_now"`
	RemindAt time.Time             `json:"remind_at" validate:"required"`
}

type UpdateReminderPayload struct {
	Type     reminder.ReminderType `json:"type" validate:"omitempty,oneof=before_due due_now"`
	RemindAt *time.Time            `json:"remind_at" validate:"omitempty"`
}

func (app *application) CreateReminderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := utils.GetUserFromContext(ctx)
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	var payload CreateReminderPayload
	if err := utils.ReadJson(w, r, &payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if _, err := app.store.Tasks.GetByID(ctx, payload.TaskID, user.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, errors.New("task not found"))
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	payload.RemindAt = payload.RemindAt.UTC()
	if payload.RemindAt.Before(time.Now().UTC()) {
		utils.BadRequestError(w, r, errors.New("remind_at must be a future date"))
		return
	}

	newReminder := &store.Reminder{
		TaskID:   payload.TaskID,
		UserID:   user.ID,
		Type:     payload.Type,
		Status:   reminder.ReminderStatusPending,
		RemindAt: payload.RemindAt,
	}

	if err := app.store.Reminders.Create(ctx, newReminder); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, errors.New("task not found"))
			return
		}
		if errors.Is(err, store.ErrConflict) {
			utils.ConflictErr(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusCreated, newReminder); err != nil {
		utils.InternalServerError(w, r, err)
	}
}

func (app *application) GetRemindersByTaskIDHandler(w http.ResponseWriter, r *http.Request) {
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

	if _, err := app.store.Tasks.GetByID(ctx, taskID, user.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, errors.New("task not found"))
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	reminders, err := app.store.Reminders.GetByTaskID(ctx, taskID, user.ID)
	if err != nil {
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusOK, reminders); err != nil {
		utils.InternalServerError(w, r, err)
	}
}

func (app *application) GetReminderByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := utils.GetUserFromContext(ctx)
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	reminderID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	existingReminder, err := app.store.Reminders.GetByID(ctx, reminderID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusOK, existingReminder); err != nil {
		utils.InternalServerError(w, r, err)
	}
}

func (app *application) UpdateReminderByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := utils.GetUserFromContext(ctx)
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	reminderID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	var payload UpdateReminderPayload
	if err := utils.ReadJson(w, r, &payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	existingReminder, err := app.store.Reminders.GetByID(ctx, reminderID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if payload.RemindAt != nil {
		remindAt := payload.RemindAt.UTC()
		if remindAt.Before(time.Now().UTC()) {
			utils.BadRequestError(w, r, errors.New("remind_at must be a future date"))
			return
		}
		payload.RemindAt = &remindAt
	}

	updatedReminder := &store.Reminder{
		ID:               existingReminder.ID,
		TaskID:           existingReminder.TaskID,
		UserID:           existingReminder.UserID,
		Type:             existingReminder.Type,
		Status:           existingReminder.Status,
		RemindAt:         existingReminder.RemindAt,
		Attempts:         existingReminder.Attempts,
		LastAttemptError: existingReminder.LastAttemptError,
		SentAt:           existingReminder.SentAt,
		CreatedAt:        existingReminder.CreatedAt,
		UpdatedAt:        existingReminder.UpdatedAt,
	}

	if payload.Type != "" {
		updatedReminder.Type = payload.Type
	}

	if payload.RemindAt != nil {
		updatedReminder.RemindAt = *payload.RemindAt
	}

	if err := app.store.Reminders.UpdateByID(ctx, reminderID, user.ID, updatedReminder); err != nil {
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

	refreshedReminder, err := app.store.Reminders.GetByID(ctx, reminderID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	if err := utils.JsonResponse(w, http.StatusOK, refreshedReminder); err != nil {
		utils.InternalServerError(w, r, err)
	}
}

func (app *application) DeleteReminderByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := utils.GetUserFromContext(ctx)
	if user == nil {
		utils.UnauthorizedError(w, r, errors.New("user not found in request context"))
		return
	}

	reminderID, err := utils.GetIDParam(r)
	if err != nil {
		utils.BadRequestError(w, r, err)
		return
	}

	if err := app.store.Reminders.DeleteByID(ctx, reminderID, user.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			utils.NotFoundError(w, r, err)
			return
		}
		utils.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
