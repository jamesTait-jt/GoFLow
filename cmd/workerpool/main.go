package main

import (
	"context"
	"fmt"

	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/cmd/workerpool/config"
	"github.com/jamesTait-jt/goflow/cmd/workerpool/plugin"
	"github.com/jamesTait-jt/goflow/cmd/workerpool/service"
	"github.com/jamesTait-jt/goflow/pkg/serialise"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/jamesTait-jt/goflow/workerpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	conf := config.LoadConfigFromFlags()

	taskQueue, resultQueue, err := initialiseBrokers(conf.BrokerType, conf.BrokerAddr)
	if err != nil {
		fmt.Println(err)

		return
	}

	taskHandlers, err := initaliseTaskHandlerRegistry(conf.HandlersPath)
	if err != nil {
		fmt.Println(err)

		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := workerpool.New(conf.NumWorkers)

	workerpoolService := service.NewWorkerpoolService(pool, taskQueue, resultQueue, taskHandlers)
	workerpoolService.Start(ctx)
}

func initialiseBrokers(brokerType, brokerAddr string) (task.Dequeuer[task.Task], task.Submitter[task.Result], error) {
	switch brokerType {
	case "redis":
		redisClient := redis.NewClient(&redis.Options{
			Addr: brokerAddr,
		})

		// TODO: Default to gob for now - allow options
		taskSerialiser := serialise.NewGobSerialiser[task.Task]()
		resultSerialiser := serialise.NewGobSerialiser[task.Result]()

		taskQueue := broker.NewRedisBroker[task.Task](redisClient, "tasks", nil, taskSerialiser)
		resultQueue := broker.NewRedisBroker[task.Result](redisClient, "results", resultSerialiser, nil)

		return taskQueue, resultQueue, nil

	default:
		return nil, nil, fmt.Errorf("invalid broker type: %s", brokerType)
	}
}

func initaliseTaskHandlerRegistry(pluginDir string) (workerpool.HandlerGetter, error) {
	plugins, err := plugin.Load(pluginDir)
	if err != nil {
		return nil, err
	}

	taskHandlers := store.NewInMemoryKVStore[string, task.Handler]()

	for pluginName, plg := range plugins {
		symbol, err := plg.Lookup("NewHandler")
		if err != nil {
			return nil, err
		}

		handlerFactory, ok := symbol.(func() task.Handler)
		if !ok {
			return nil, fmt.Errorf("invalid plugin: Handler does not implement Handler interface")
		}

		handler := handlerFactory()

		taskHandlers.Put(pluginName, handler)
	}

	return taskHandlers, nil
}
