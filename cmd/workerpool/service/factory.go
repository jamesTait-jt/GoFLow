package service

import (
	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/jamesTait-jt/goflow/workerpool"
	"github.com/redis/go-redis/v9"
)

type Factory struct {
	pool         workerpoolRunner
	serialiser   broker.Serialiser[task.Result]
	deserialiser broker.Deserialiser[task.Task]
	taskHandlers workerpool.HandlerGetter
}

func NewFactory(
	pool workerpoolRunner,
	serialiser broker.Serialiser[task.Result],
	deserialiser broker.Deserialiser[task.Task],
	taskHandlers workerpool.HandlerGetter,
) *Factory {
	return &Factory{
		pool:         pool,
		serialiser:   serialiser,
		deserialiser: deserialiser,
		taskHandlers: taskHandlers,
	}
}

func (f *Factory) CreateRedisWorkerpoolService(client *redis.Client) *WorkerpoolService {
	taskQueue := broker.NewRedisBroker(client, "tasks", nil, f.deserialiser)
	resultQueue := broker.NewRedisBroker(client, "results", f.serialiser, nil)

	return NewWorkerpoolService(f.pool, taskQueue, resultQueue, f.taskHandlers)
}
