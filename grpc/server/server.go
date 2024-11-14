// Package server provides the implementation of a gRPC server for the GoFlow service.
// It includes a server that handles task submission and result retrieval via gRPC methods.
//
// The GoFlowGRPCServer is configured with various options such as the logger and port,
// and can be started and stopped gracefully. The GoFlowService implements the core
// logic for handling tasks, interfacing with the GoFlow task execution system.
//
// The package also defines a controller for the GoFlow gRPC service, GoFlowServiceController,
// which processes gRPC requests such as pushing tasks and fetching results.
//
// This package uses the GoFlow service's task management capabilities and integrates
// with gRPC to provide an interface for remote task submission and result retrieval.
package server

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

// GoFlowGRPCServer represents a gRPC server that handles GoFlow service requests.
// It encapsulates the gRPC server instance and configuration options such as
// the logger and port number. This struct provides methods to start and stop
// the server as well as manage gRPC service registration.
//
// The server listens for incoming requests on the configured port and uses
// the provided logger for reporting events during server operation.
type GoFlowGRPCServer struct {
	opts       goFlowGRPCServerOptions
	grpcServer *grpc.Server
}

// New creates a new GoFlowGRPCServer with the specified options.
// It initializes the gRPC server and applies any provided configuration options
// (such as logger and port) to the server's options.
// The default options are used if no options are provided.
//
// TODO: Expand this to provide access to set gRPC options
func New(opt ...GoFlowGRPCServerOption) *GoFlowGRPCServer {
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

// Start starts the gRPC server and listens for incoming connections.
// It binds to the port specified in the server options and registers the provided
// service with the gRPC server. Once started, it enters a blocking state and begins
// serving requests until the server is gracefully stopped.
//
// Example usage:
//
//	grpcServer := server.New(
//	    server.WithPort(50051),
//	    server.WithLogger(log.Default()),
//	)
//
//	go grpcServer.Start(
//		func(server *grpc.Server) {
//			pb.RegisterGoFlowServer(server, controller)
//		},
//	)
//
// In this example, the server starts listening on port 50051, and the provided
// service (`GoFlowService`) is registered with the gRPC server.
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

// Close gracefully shuts down the GoFlow gRPC server. It logs the closure
// event and stops the server, ensuring that any in-progress requests
// are completed before the server is fully closed.
func (g GoFlowGRPCServer) Close() error {
	g.opts.logger.Info("closing gRPC server")

	g.grpcServer.GracefulStop()

	return nil
}
