package broker

import (
	"github.com/jamesTait-jt/goflow/pkg/log"
)

type redisBrokerOptions struct {
	logger log.Logger
}

func defaultRedisBrokerOptions() redisBrokerOptions {
	return redisBrokerOptions{
		logger: log.NewConsoleLogger(),
	}
}

// A RedisBrokerOption sets options such as logger.
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
