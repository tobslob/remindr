package reminder

import (
	"context"
	"log"
	"time"
)

type SchedulerStore interface {
	ClaimDue(context.Context, int64) ([]*Record, error)
}

type Scheduler struct {
	store    SchedulerStore
	jobs     chan<- Job
	interval time.Duration
	batch    int64
	logger   *log.Logger
}

type SchedulerConfig struct {
	Interval time.Duration
	Batch    int64
	Logger   *log.Logger
}

func NewScheduler(store SchedulerStore, jobs chan<- Job, cfg SchedulerConfig) *Scheduler {
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	if cfg.Batch <= 0 {
		cfg.Batch = 25
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}

	return &Scheduler{
		store:    store,
		jobs:     jobs,
		interval: cfg.Interval,
		batch:    cfg.Batch,
		logger:   cfg.Logger,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	s.claimAndEnqueue(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.claimAndEnqueue(ctx)
		}
	}
}

func (s *Scheduler) claimAndEnqueue(ctx context.Context) {
	reminders, err := s.store.ClaimDue(ctx, s.batch)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		s.logger.Printf("claim due reminders: %v", err)
		return
	}

	for _, r := range reminders {
		job := Job{ReminderID: r.ID}
		select {
		case s.jobs <- job:
		case <-ctx.Done():
			return
		}
	}
}
