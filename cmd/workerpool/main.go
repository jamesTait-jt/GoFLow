package main

import (
	"context"
	"fmt"
	"plugin"

	"github.com/jamesTait-jt/goflow/cmd/workerpool/config"
	"github.com/jamesTait-jt/goflow/cmd/workerpool/pluginloader"
	"github.com/jamesTait-jt/goflow/cmd/workerpool/service"
	"github.com/jamesTait-jt/goflow/cmd/workerpool/taskhandlers"
	"github.com/jamesTait-jt/goflow/pkg/serialise"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/jamesTait-jt/goflow/workerpool"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/afero"
)

func main() {
	conf := config.LoadConfigFromFlags()

	pool := workerpool.New(conf.NumWorkers)

	pluginLoader := pluginloader.New(afero.NewOsFs(), plugin.Open)

	taskHandlers, err := taskhandlers.Load(pluginLoader, conf.HandlersPath)
	if err != nil {
		fmt.Println(err)

		return
	}

	resultSerialiser := serialise.NewGobSerialiser[task.Result]()
	taskSerialiser := serialise.NewGobSerialiser[task.Task]()
	serviceFactory := service.NewFactory(pool, resultSerialiser, taskSerialiser, taskHandlers)

	var workerpoolService *service.WorkerpoolService

	switch conf.BrokerType {
	case "redis":
		client := redis.NewClient(&redis.Options{
			Addr: conf.BrokerAddr,
		})
		workerpoolService = serviceFactory.CreateRedisWorkerpoolService(client)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerpoolService.Start(ctx)
}
