package runtime

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

type Runner struct {
	conf *config.Config
}

func New() *Runner {
	return &Runner{
		conf: config.LoadConfigFromFlags(),
	}
}

func (r *Runner) Run() {
	pool := workerpool.New(r.conf.NumWorkers)

	pluginLoader := pluginloader.New(afero.NewOsFs(), plugin.Open)

	taskHandlers, err := taskhandlers.Load(pluginLoader, r.conf.HandlersPath)
	if err != nil {
		fmt.Println(err)

		return
	}

	resultSerialiser := serialise.NewGobSerialiser[task.Result]()
	taskSerialiser := serialise.NewGobSerialiser[task.Task]()
	serviceFactory := service.NewFactory(pool, resultSerialiser, taskSerialiser, taskHandlers)

	var workerpoolService *service.WorkerpoolService

	switch r.conf.BrokerType {
	case "redis":
		client := redis.NewClient(&redis.Options{
			Addr: r.conf.BrokerAddr,
		})
		workerpoolService = serviceFactory.CreateRedisWorkerpoolService(client)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerpoolService.Start(ctx)
}
