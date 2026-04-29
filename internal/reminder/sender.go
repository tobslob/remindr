package reminder

import (
	"context"
	"log"
)

type Sender interface {
	Send(context.Context, *Record) error
}

type SenderFunc func(context.Context, *Record) error

func (f SenderFunc) Send(ctx context.Context, r *Record) error {
	return f(ctx, r)
}

type LogSender struct {
	logger *log.Logger
}

func NewLogSender(logger *log.Logger) *LogSender {
	if logger == nil {
		logger = log.Default()
	}

	return &LogSender{logger: logger}
}

func (s *LogSender) Send(_ context.Context, r *Record) error {
	s.logger.Printf(
		"reminder due: reminder_id=%s task_id=%s user_id=%s type=%s remind_at=%s",
		r.ID,
		r.TaskID,
		r.UserID,
		r.Type,
		r.RemindAt.Format("2006-01-02T15:04:05Z07:00"),
	)

	return nil
}
