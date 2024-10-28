package kubernetes

import "github.com/jamesTait-jt/goflow/pkg/log"

type Option interface {
	apply(*options)
}

type options struct {
	configBuilder     kubeConfigBuilder
	kubeClientBuilder kubeClientBuilder
	logger            log.Logger
}

func defaultOptions() options {
	return options{
		configBuilder:     &KubeConfigBuilder{},
		kubeClientBuilder: &KubeClientBuilder{},
		logger:            log.NewConsoleLogger(),
	}
}

type configBuilderOption struct {
	ConfigBuilder kubeConfigBuilder
}

func (c configBuilderOption) apply(opts *options) {
	opts.configBuilder = c.ConfigBuilder
}

func WithConfigBuilder(configBuilder kubeConfigBuilder) Option {
	return configBuilderOption{ConfigBuilder: configBuilder}
}

type kubeClientBuilderOption struct {
	KubeClientBuilder kubeClientBuilder
}

func (k kubeClientBuilderOption) apply(opts *options) {
	opts.kubeClientBuilder = k.KubeClientBuilder
}

func WithKubeClientBuilder(kubeClientBuilder kubeClientBuilder) Option {
	return kubeClientBuilderOption{KubeClientBuilder: kubeClientBuilder}
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
