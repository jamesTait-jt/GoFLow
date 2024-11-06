package service

import (
	"context"

	"github.com/jamesTait-jt/goflow/task"
	"github.com/jamesTait-jt/goflow/workerpool"
)

type workerpoolRunner interface {
	Start(ctx context.Context, taskQueue task.Dequeuer[task.Task], results task.Submitter[task.Result], taskHandlers workerpool.HandlerGetter)
	AwaitShutdown()
}

type WorkerpoolService struct {
	pool         workerpoolRunner
	taskQueue    task.Dequeuer[task.Task]
	resultQueue  task.Submitter[task.Result]
	taskHandlers workerpool.HandlerGetter
}

func NewWorkerpoolService(
	pool workerpoolRunner,
	taskQueue task.Dequeuer[task.Task],
	resultQueue task.Submitter[task.Result],
	taskHandlers workerpool.HandlerGetter,
) *WorkerpoolService {
	return &WorkerpoolService{
		pool:         pool,
		taskQueue:    taskQueue,
		resultQueue:  resultQueue,
		taskHandlers: taskHandlers,
	}
}

func (w *WorkerpoolService) Start(ctx context.Context) {
	w.pool.Start(ctx, w.taskQueue, w.resultQueue, w.taskHandlers)
	w.pool.AwaitShutdown()
}
