package utils

import (
	"context"
	"log"
	"sync"
	"time"
)

// rannable task interface
type Task interface {
	Run(ctx context.Context)
}

type TaskOptFunc func(*TaskRunner)

// task scheduler
type TaskRunner struct {
	runOnStart bool
	interval   time.Duration
	timeout    time.Duration

	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc

	handler Task

	running   bool
	executing bool
	mu        *sync.Mutex
}

// new taskRunner
func NewTaskRunner(handler Task, opts ...TaskOptFunc) *TaskRunner {
	taskRunner := &TaskRunner{
		runOnStart: true,
		interval:   1 * time.Hour,
		timeout:    5 * time.Minute,
		handler:    handler,

		running: false,
		mu:      &sync.Mutex{},
	}

	for _, opt := range opts {
		opt(taskRunner)
	}

	return taskRunner
}

func WithRunOnStart(run bool) TaskOptFunc {
	return func(tr *TaskRunner) {
		tr.runOnStart = run
	}
}

// SetInterval 设置任务执行间隔
func WithInterval(interval time.Duration) TaskOptFunc {
	return func(tr *TaskRunner) {
		tr.interval = interval
	}
}

// set task timeout
func WithTimeout(timeout time.Duration) TaskOptFunc {
	return func(tr *TaskRunner) {
		tr.timeout = timeout
	}
}

// start scheduler
func (s *TaskRunner) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		log.Printf("Scheduler is running, skip")
		return
	}

	s.ticker = time.NewTicker(s.interval)
	s.ctx, s.cancel = context.WithCancel(ctx)

	s.running = true

	log.Printf("scheduler started, interval: %v", s.interval)

	if s.runOnStart {
		var tCtx context.Context
		var tCancel context.CancelFunc
		if s.timeout > 0 {
			tCtx, tCancel = context.WithTimeout(s.ctx, s.timeout)
		} else {
			tCtx, tCancel = context.WithCancel(s.ctx)
		}
		go s.ExecuteOnce(tCtx, tCancel)
	}

	go s.runLoop()
}

// stop scheduler
func (s *TaskRunner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		log.Printf("scheduler not running, no need to stop")
		return
	}

	s.ticker.Stop()
	if s.cancel != nil {
		s.cancel()
	}

	s.running = false
	log.Printf("scheduler stopped")
}

func (s *TaskRunner) runLoop() {
	for {
		select {
		case <-s.ticker.C:
			s.mu.Lock()
			if s.executing {
				log.Printf("skip")
				s.mu.Unlock()
				continue
			}
			s.executing = true
			s.mu.Unlock()

			var tCtx context.Context
			var tCancel context.CancelFunc
			if s.timeout > 0 {
				tCtx, tCancel = context.WithTimeout(s.ctx, s.timeout)
			} else {
				tCtx, tCancel = context.WithCancel(s.ctx)
			}

			log.Printf("task start...")
			go s.ExecuteOnce(tCtx, tCancel)

		case <-s.ctx.Done():
			log.Println("scheduler received cancel signal, stop...")
			s.ticker.Stop()
			return
		}
	}
}

// sync excute task once
func (s *TaskRunner) ExecuteOnce(ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		if cancel != nil {
			cancel()
		}
		if r := recover(); r != nil {
			log.Printf("panic occurred during task: %v", r)
		}
		s.mu.Lock()
		s.executing = false
		s.mu.Unlock()
	}()

	s.handler.Run(ctx)
}
