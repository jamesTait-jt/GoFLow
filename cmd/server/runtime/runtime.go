package runtime

import (
	"context"
	"fmt"

	"github.com/jamesTait-jt/goflow"
	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/cmd/server/config"
	pb "github.com/jamesTait-jt/goflow/grpc/proto"
	"github.com/jamesTait-jt/goflow/grpc/server"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/pkg/serialise"
	"github.com/jamesTait-jt/goflow/pkg/shutdown"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type Runtime struct {
	Conf *config.Config
}

func New() *Runtime {
	return &Runtime{Conf: config.LoadConfigFromFlags()}
}

func (r *Runtime) Run(ctx context.Context) error {
	logger := log.NewConsoleLogger()

	redisClient := redis.NewClient(&redis.Options{
		Addr: r.Conf.BrokerAddr,
	})

	pong, err := redisClient.Ping(ctx).Result()

	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("redis connection successful: %s", pong))

	taskSubmitter := broker.NewRedisBroker(redisClient, "tasks", serialise.NewGobSerialiser[task.Task](), nil, logger)
	resultsGetter := broker.NewRedisBroker(redisClient, "results", nil, serialise.NewGobSerialiser[task.Result](), logger)
	resultsStore := store.NewInMemoryKVStore[string, task.Result]()

	gf := goflow.New(
		taskSubmitter,
		resultsGetter,
		goflow.WithResultsStore(resultsStore),
	)

	_ = gf.Start()

	gfService := server.NewGoFlowService(gf)
	controller := server.NewGoFlowServiceController(gfService, logger)

	grpcServer := server.New(logger)

	go grpcServer.Start(
		func(server *grpc.Server) {
			pb.RegisterGoFlowServer(server, controller)
		},
	)

	shutdown.AddShutdownHook(ctx, logger, grpcServer, redisClient, gf)

	return nil
}
