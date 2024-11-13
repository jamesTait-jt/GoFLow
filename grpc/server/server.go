package server

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type GoFlowGRPCServer struct {
	opts       goFlowGRPCServerOptions
	grpcServer *grpc.Server
}

// TODO: Expand this to provide access to set gRPC options
func New(opt ...GoFlowServerOption) *GoFlowGRPCServer {
	opts := defaultServerOptions

	for _, o := range opt {
		o.apply(&opts)
	}

	s := grpc.NewServer()

	return &GoFlowGRPCServer{
		grpcServer: s,
		opts:       opts,
	}
}

func (g GoFlowGRPCServer) Start(serviceRegister func(server *grpc.Server)) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", g.opts.port))
	if err != nil {
		g.opts.logger.Fatal(fmt.Sprintf("failed to start gRPC server: %v", err))
	}

	serviceRegister(g.grpcServer)

	g.opts.logger.Info(fmt.Sprintf("server listening at %v", lis.Addr()))

	if err := g.grpcServer.Serve(lis); err != nil {
		g.opts.logger.Fatal(fmt.Sprintf("failed to serve gRPC server: %v", err))
	}
}

func (g GoFlowGRPCServer) Close() error {
	g.opts.logger.Info("closing gRPC server")

	g.grpcServer.GracefulStop()

	return nil
}
