package resource

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedapiv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Namespace struct {
	config *acapiv1.NamespaceApplyConfiguration
	client typedapiv1.NamespaceInterface
}

func NewNamespace(config *acapiv1.NamespaceApplyConfiguration, client typedapiv1.NamespaceInterface) *Namespace {
	return &Namespace{
		config: config,
		client: client,
	}
}

func (n *Namespace) Name() string {
	return *n.config.Name
}

func (n *Namespace) Kind() string {
	return "namespace"
}

func (n *Namespace) Apply(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
	return n.client.Apply(ctx, n.config, opts)
}

func (n *Namespace) Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
	return n.client.Get(ctx, n.Name(), opts)
}
