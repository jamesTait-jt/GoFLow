package broker

import (
	"time"

	"github.com/jamesTait-jt/goflow/pkg/log"
)

type redisBrokerOptions struct {
	pollTimeout time.Duration
	logger      log.Logger
}

func defaultRedisBrokerOptions() redisBrokerOptions {
	return redisBrokerOptions{
		pollTimeout: time.Second,
		logger:      log.NewConsoleLogger(),
	}
}

// A RedisBrokerOption sets options such as logger, poll delay, etc.
type RedisBrokerOption interface {
	apply(*redisBrokerOptions)
}

type loggerOption struct {
	Logger log.Logger
}

func (l loggerOption) apply(opts *redisBrokerOptions) {
	opts.logger = l.Logger
}

// WithLogger allows you to set logger that will report on basic warnings when
// interacting with redis.
func WithLogger(logger log.Logger) RedisBrokerOption {
	return loggerOption{Logger: logger}
}

type pollTimeoutOption struct {
	PollTimeout time.Duration
}

func (p pollTimeoutOption) apply(opts *redisBrokerOptions) {
	opts.pollTimeout = p.PollTimeout
}

// WithPollTimeout sets the timeout duration for each Redis BRPop operation.
// This timeout specifies how long the broker should wait for tasks to become
// available in the Redis queue before timing out and checking for shutdown signals.
// A shorter timeout increases responsiveness to shutdown.
func WithPollTimeout(timeout time.Duration) RedisBrokerOption {
	return pollTimeoutOption{PollTimeout: timeout}
}
