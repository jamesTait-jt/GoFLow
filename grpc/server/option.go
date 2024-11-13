package server

import (
	"github.com/jamesTait-jt/goflow/pkg/log"
)

type goFlowGRPCServerOptions struct {
	logger log.Logger
	port   int
}

var (
	defaultgRPCPort = 50051

	defaultServerOptions = goFlowGRPCServerOptions{
		logger: log.NewConsoleLogger(),
		port:   defaultgRPCPort,
	}
)

// A GoFlowGRPCServerOption sets options such as logger, port, etc.
type GoFlowGRPCServerOption interface {
	apply(*goFlowGRPCServerOptions)
}

type loggerOption struct {
	Logger log.Logger
}

func (l loggerOption) apply(opts *goFlowGRPCServerOptions) {
	opts.logger = l.Logger
}

// WithLogger allows you to set logger that will report on basic server start/stop
// operations.
func WithLogger(logger log.Logger) GoFlowGRPCServerOption {
	return loggerOption{Logger: logger}
}

type portOption struct {
	Port int
}

func (p portOption) apply(opts *goFlowGRPCServerOptions) {
	opts.port = p.Port
}

// WithPort allows you to set the port on which the server will listen.
func WithPort(port int) GoFlowGRPCServerOption {
	return portOption{Port: port}
}
