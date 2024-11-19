package k8s

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
