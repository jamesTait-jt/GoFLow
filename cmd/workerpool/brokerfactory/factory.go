package brokerfactory

import (
	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/redis/go-redis/v9"
)

type Factory struct {
	serialiser   broker.Serialiser[task.Result]
	deserialiser broker.Deserialiser[task.Task]
}

func NewFactory(s broker.Serialiser[task.Result], d broker.Deserialiser[task.Task]) *Factory {
	return &Factory{
		serialiser:   s,
		deserialiser: d,
	}
}

func (f *Factory) CreateRedisBrokers(client *redis.Client) (tasks *broker.RedisBroker[task.Task], results *broker.RedisBroker[task.Result]) {
	taskQueue := broker.NewRedisBroker[task.Task](client, "tasks", nil, f.deserialiser)
	resultQueue := broker.NewRedisBroker[task.Result](client, "results", f.serialiser, nil)

	return taskQueue, resultQueue
}
