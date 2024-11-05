package run

import (
	"time"

	"github.com/jamesTait-jt/goflow/pkg/log"
)

type Option interface {
	apply(*options)
}

type options struct {
	logger  log.Logger
	timeout time.Duration
}

func defaultOptions() options {
	return options{
		logger:  log.NewConsoleLogger(),
		timeout: 30 * time.Second,
	}
}

type loggerOption struct {
	Logger log.Logger
}

func (l loggerOption) apply(opts *options) {
	opts.logger = l.Logger
}

func WithLogger(logger log.Logger) Option {
	return loggerOption{Logger: logger}
}

type timeoutOption struct {
	Timeout time.Duration
}

func (l timeoutOption) apply(opts *options) {
	opts.timeout = l.Timeout
}

func WithTimeout(timeout time.Duration) Option {
	return timeoutOption{Timeout: timeout}
}
