package runtime

import (
	"context"
	"fmt"
	"plugin"

	"github.com/jamesTait-jt/goflow/cmd/workerpool/config"
	"github.com/jamesTait-jt/goflow/cmd/workerpool/pluginloader"
	"github.com/jamesTait-jt/goflow/cmd/workerpool/service"
	"github.com/jamesTait-jt/goflow/cmd/workerpool/taskhandlers"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/pkg/serialise"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/jamesTait-jt/goflow/workerpool"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/afero"
)

type Runtime struct {
	Conf *config.Config
}

func New() *Runtime {
	return &Runtime{
		Conf: config.LoadConfigFromFlags(),
	}
}

func (r *Runtime) Run() error {
	fmt.Printf("workerpool started with config: %v\n", r.Conf)
	pool := workerpool.New(r.Conf.NumWorkers)

	pluginLoader := pluginloader.New(afero.NewOsFs(), plugin.Open)

	taskHandlers, err := taskhandlers.Load(pluginLoader, r.Conf.HandlersPath)
	if err != nil {
		return err
	}

	logger := log.NewConsoleLogger()

	resultSerialiser := serialise.NewGobSerialiser[task.Result]()
	taskSerialiser := serialise.NewGobSerialiser[task.Task]()
	serviceFactory := service.NewFactory(pool, taskSerialiser, resultSerialiser, taskHandlers, logger)

	var workerpoolService *service.WorkerpoolService

	switch r.Conf.BrokerType {
	case "redis":
		client := redis.NewClient(&redis.Options{
			Addr: r.Conf.BrokerAddr,
		})
		ctx := context.Background()
		pong, err := client.Ping(ctx).Result()

		if err != nil {
			return fmt.Errorf("could not connect to redis: %v", err)
		}

		fmt.Printf("redis connection successful: %s\n", pong)

		workerpoolService = serviceFactory.CreateRedisWorkerpoolService(client)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerpoolService.Start(ctx)

	return nil
}
