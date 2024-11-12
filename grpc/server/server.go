package server

import (
	"fmt"
	"net"

	"github.com/jamesTait-jt/goflow/pkg/log"
	"google.golang.org/grpc"
)

type GoFlowGRPCServer struct {
	grpcServer *grpc.Server
	logger     log.Logger
}

func New(logger log.Logger) *GoFlowGRPCServer {
	s := grpc.NewServer()

	return &GoFlowGRPCServer{
		grpcServer: s,
		logger:     logger,
	}
}

func (g GoFlowGRPCServer) Start(serviceRegister func(server *grpc.Server)) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051))
	if err != nil {
		g.logger.Fatal(fmt.Sprintf("failed to start gRPC server: %v", err))
	}

	serviceRegister(g.grpcServer)

	g.logger.Info(fmt.Sprintf("server listening at %v", lis.Addr()))

	if err := g.grpcServer.Serve(lis); err != nil {
		g.logger.Fatal(fmt.Sprintf("failed to serve gRPC server: %v", err))
	}
}

func (g GoFlowGRPCServer) Close() error {
	g.logger.Info("closing gRPC server")

	g.grpcServer.GracefulStop()

	return nil
}
