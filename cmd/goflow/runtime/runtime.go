package runtime

import (
	"context"
	"fmt"
	"net"

	"github.com/jamesTait-jt/goflow"
	"github.com/jamesTait-jt/goflow/broker"
	pb "github.com/jamesTait-jt/goflow/grpc/proto"
	"github.com/jamesTait-jt/goflow/grpc/server/controller"
	"github.com/jamesTait-jt/goflow/grpc/server/service"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/pkg/serialise"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

var (
	redisPort  = "6379"
	serverPort = "50051"
)

type Runtime struct{}

func New() *Runtime {
	return &Runtime{}
}

func (r *Runtime) Run() error {
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("goflow-redis-server:%s", redisPort),
	})
	ctx := context.Background()
	pong, err := redisClient.Ping(ctx).Result()

	if err != nil {
		return err
	}

	logger := log.NewConsoleLogger()

	logger.Info(fmt.Sprintf("redis connection successful: %s", pong))

	taskSubmitter := broker.NewRedisBroker(redisClient, "tasks", serialise.NewGobSerialiser[task.Task](), nil, logger)
	resultsGetter := broker.NewRedisBroker(redisClient, "results", nil, serialise.NewGobSerialiser[task.Result](), logger)
	resultsStore := store.NewInMemoryKVStore[string, task.Result]()

	gf := goflow.New(
		taskSubmitter,
		resultsGetter,
		goflow.WithResultsStore(resultsStore),
	)

	gf.Start()

	gfService := service.New(gf)
	controller := controller.NewGoFlowServiceController(gfService, logger)

	grpcServer := grpc.NewServer()
	pb.RegisterGoFlowServer(grpcServer, controller)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", serverPort))
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("server listening at %v", lis.Addr()))

	return grpcServer.Serve(lis)
}
