package service

import (
	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/jamesTait-jt/goflow/workerpool"
	"github.com/redis/go-redis/v9"
)

type Factory struct {
	pool          workerpoolRunner
	taskEncoder   broker.Encoder[task.Task]
	resultEncoder broker.Encoder[task.Result]
	taskHandlers  workerpool.HandlerGetter
	logger        log.Logger
}

func NewFactory(
	pool workerpoolRunner,
	taskEncoder broker.Encoder[task.Task],
	resultEncoder broker.Encoder[task.Result],
	taskHandlers workerpool.HandlerGetter,
	logger log.Logger,
) *Factory {
	return &Factory{
		pool:          pool,
		taskEncoder:   taskEncoder,
		resultEncoder: resultEncoder,
		taskHandlers:  taskHandlers,
		logger:        logger,
	}
}

func (f *Factory) CreateRedisWorkerpoolService(client *redis.Client) *WorkerpoolService {
	taskQueue := broker.NewRedisBroker(client, "tasks", f.taskEncoder, broker.WithLogger(f.logger))
	resultQueue := broker.NewRedisBroker(client, "results", f.resultEncoder, broker.WithLogger(f.logger))

	return NewWorkerpoolService(f.pool, taskQueue, resultQueue, f.taskHandlers)
}
