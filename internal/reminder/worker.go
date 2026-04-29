package reminder

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
)

type WorkerStore interface {
	GetForSendingByID(context.Context, uuid.UUID) (*Record, error)
	MarkSentByID(context.Context, uuid.UUID) error
	MarkFailedByID(context.Context, uuid.UUID, string) error
}

type Worker struct {
	store  WorkerStore
	sender Sender
	jobs   <-chan Job
	logger *log.Logger
}

func NewWorker(store WorkerStore, sender Sender, jobs <-chan Job, logger *log.Logger) *Worker {
	if logger == nil {
		logger = log.Default()
	}

	return &Worker{
		store:  store,
		sender: sender,
		jobs:   jobs,
		logger: logger,
	}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-w.jobs:
			w.Handle(ctx, job)
		}
	}
}

func (w *Worker) Handle(ctx context.Context, job Job) {
	r, err := w.store.GetForSendingByID(ctx, job.ReminderID)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		w.logger.Printf("get reminder for sending: reminder_id=%s error=%v", job.ReminderID, err)
		return
	}

	if err := w.sender.Send(ctx, r); err != nil {
		if markErr := w.store.MarkFailedByID(ctx, r.ID, err.Error()); markErr != nil && !errors.Is(markErr, context.Canceled) {
			w.logger.Printf("mark reminder failed: reminder_id=%s send_error=%v mark_error=%v", r.ID, err, markErr)
		}
		return
	}

	if err := w.store.MarkSentByID(ctx, r.ID); err != nil && ctx.Err() == nil {
		w.logger.Printf("mark reminder sent: reminder_id=%s error=%v", r.ID, err)
	}
}
