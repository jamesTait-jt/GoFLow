package client

import (
	"time"

	"github.com/jamesTait-jt/goflow/pkg/log"
)

type goFlowGRPCClientOptions struct {
	logger         log.Logger
	requestTimeout time.Duration
}

var (
	defaultRequestTimeout = 30 * time.Second

	defaultServerOptions = goFlowGRPCClientOptions{
		logger:         log.NewConsoleLogger(),
		requestTimeout: defaultRequestTimeout,
	}
)

// A GoFlowGRPCClientOption sets options such as logger, request timeout, etc.
type GoFlowGRPCClientOption interface {
	apply(*goFlowGRPCClientOptions)
}

type loggerOption struct {
	Logger log.Logger
}

func (l loggerOption) apply(opts *goFlowGRPCClientOptions) {
	opts.logger = l.Logger
}

// WithLogger allows you to set logger that will report on basic server start/stop
// operations.
func WithLogger(logger log.Logger) GoFlowGRPCClientOption {
	return loggerOption{Logger: logger}
}

type requestTimeoutOption struct {
	RequestTimeout time.Duration
}

func (p requestTimeoutOption) apply(opts *goFlowGRPCClientOptions) {
	opts.requestTimeout = p.RequestTimeout
}

// WithRequestTimeout allows you to set the time the client will wait after sending a
// request before timing out.
func WithRequestTimeout(requestTimeout time.Duration) GoFlowGRPCClientOption {
	return requestTimeoutOption{RequestTimeout: requestTimeout}
}
