package kubernetes

import "github.com/jamesTait-jt/goflow/pkg/log"

type OperatorOption interface {
	apply(*operatorOptions)
}

type operatorOptions struct {
	logger log.Logger
}

func defaultOperatorOptions() operatorOptions {
	return operatorOptions{
		// configBuilder:     &KubeConfigBuilder{},
		// kubeClientBuilder: &KubeClientBuilder{},
		logger: log.NewConsoleLogger(),
	}
}

type loggerOption struct {
	Logger log.Logger
}

func (l loggerOption) apply(opts *operatorOptions) {
	opts.logger = l.Logger
}

func WithLogger(logger log.Logger) OperatorOption {
	return loggerOption{Logger: logger}
}

type BuildClientsetOption interface {
	apply(*buildClientsetOptions)
}

type buildClientsetOptions struct {
	configBuilder     kubeConfigBuilder
	kubeClientBuilder clientSetBuilder
}

func defaultBuildClientsetOptions() buildClientsetOptions {
	return buildClientsetOptions{
		configBuilder:     &KubeConfigBuilder{},
		kubeClientBuilder: &KubeClientBuilder{},
	}
}

type configBuilderOption struct {
	ConfigBuilder kubeConfigBuilder
}

func (c configBuilderOption) apply(opts *buildClientsetOptions) {
	opts.configBuilder = c.ConfigBuilder
}

func WithConfigBuilder(configBuilder kubeConfigBuilder) BuildClientsetOption {
	return configBuilderOption{ConfigBuilder: configBuilder}
}

type kubeClientBuilderOption struct {
	KubeClientBuilder clientSetBuilder
}

func (k kubeClientBuilderOption) apply(opts *buildClientsetOptions) {
	opts.kubeClientBuilder = k.KubeClientBuilder
}

func WithKubeClientBuilder(kubeClientBuilder clientSetBuilder) BuildClientsetOption {
	return kubeClientBuilderOption{KubeClientBuilder: kubeClientBuilder}
}
