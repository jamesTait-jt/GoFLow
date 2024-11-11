package service

import (
	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/jamesTait-jt/goflow/workerpool"
	"github.com/redis/go-redis/v9"
)

type Factory struct {
	pool         workerpoolRunner
	serialiser   broker.Serialiser[task.Result]
	deserialiser broker.Deserialiser[task.Task]
	taskHandlers workerpool.HandlerGetter
	logger       log.Logger
}

func NewFactory(
	pool workerpoolRunner,
	serialiser broker.Serialiser[task.Result],
	deserialiser broker.Deserialiser[task.Task],
	taskHandlers workerpool.HandlerGetter,
	logger log.Logger,
) *Factory {
	return &Factory{
		pool:         pool,
		serialiser:   serialiser,
		deserialiser: deserialiser,
		taskHandlers: taskHandlers,
		logger:       logger,
	}
}

func (f *Factory) CreateRedisWorkerpoolService(client *redis.Client) *WorkerpoolService {
	taskQueue := broker.NewRedisBroker(client, "tasks", nil, f.deserialiser, f.logger)
	resultQueue := broker.NewRedisBroker(client, "results", f.serialiser, nil, f.logger)

	return NewWorkerpoolService(f.pool, taskQueue, resultQueue, f.taskHandlers)
}
