package reminder

import (
	"context"
	"log"
	"sync"
	"time"
)

type Store interface {
	SchedulerStore
	WorkerStore
}

type Service struct {
	store       Store
	sender      Sender
	logger      *log.Logger
	interval    time.Duration
	batch       int64
	workerCount int
	queueSize   int

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type ServiceConfig struct {
	Interval    time.Duration
	Batch       int64
	WorkerCount int
	QueueSize   int
	Logger      *log.Logger
}

func NewService(store Store, sender Sender, cfg ServiceConfig) *Service {
	if sender == nil {
		sender = NewLogSender(cfg.Logger)
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	if cfg.Batch <= 0 {
		cfg.Batch = 25
	}
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 2
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = int(cfg.Batch)
	}

	return &Service{
		store:       store,
		sender:      sender,
		logger:      cfg.Logger,
		interval:    cfg.Interval,
		batch:       cfg.Batch,
		workerCount: cfg.WorkerCount,
		queueSize:   cfg.QueueSize,
	}
}

func (s *Service) Start(ctx context.Context) {
	if s.cancel != nil {
		return
	}

	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	jobs := make(chan Job, s.queueSize)

	scheduler := NewScheduler(s.store, jobs, SchedulerConfig{
		Interval: s.interval,
		Batch:    s.batch,
		Logger:   s.logger,
	})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		scheduler.Run(runCtx)
	}()

	for i := 0; i < s.workerCount; i++ {
		worker := NewWorker(s.store, s.sender, jobs, s.logger)

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			worker.Run(runCtx)
		}()
	}

	s.logger.Printf("reminder service started: interval=%s batch=%d workers=%d", s.interval, s.batch, s.workerCount)
}

func (s *Service) Stop() {
	if s.cancel == nil {
		return
	}

	s.cancel()
	s.wg.Wait()
	s.cancel = nil
	s.logger.Println("reminder service stopped")
}
