package resource

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"

	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	typedapiv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Resource struct {
	name      string
	kind      string
	applyFunc func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error)
	getFunc   func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error)
}

func (r *Resource) Name() string {
	return r.name
}

func (r *Resource) Kind() string {
	return r.kind
}

func (r *Resource) Apply(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
	return r.applyFunc(ctx, opts)
}

func (r *Resource) Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
	return r.getFunc(ctx, opts)
}

type namespaceInterface interface {
	Apply(ctx context.Context, namespace *acapiv1.NamespaceApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Namespace, err error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiv1.Namespace, error)
}

func NewNamespace(config *acapiv1.NamespaceApplyConfiguration, client namespaceInterface) *Resource {
	return &Resource{
		name: *config.Name,
		kind: "namespace",
		applyFunc: func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
			return client.Apply(ctx, config, opts)
		},
		getFunc: func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
			return client.Get(ctx, *config.Name, opts)
		},
	}
}

func NewDeployment(config *acappsv1.DeploymentApplyConfiguration, client typedappsv1.DeploymentInterface) *Resource {
	return &Resource{
		name: *config.Name,
		kind: "deployment",
		applyFunc: func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
			return client.Apply(ctx, config, opts)
		},
		getFunc: func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
			return client.Get(ctx, *config.Name, opts)
		},
	}
}

func NewService(config *acapiv1.ServiceApplyConfiguration, client typedapiv1.ServiceInterface) *Resource {
	return &Resource{
		name: *config.Name,
		kind: "service",
		applyFunc: func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
			return client.Apply(ctx, config, opts)
		},
		getFunc: func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
			return client.Get(ctx, *config.Name, opts)
		},
	}
}

func NewPersistentVolume(config *acapiv1.PersistentVolumeApplyConfiguration, client typedapiv1.PersistentVolumeInterface) *Resource {
	return &Resource{
		name: *config.Name,
		kind: "pv",
		applyFunc: func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
			return client.Apply(ctx, config, opts)
		},
		getFunc: func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
			return client.Get(ctx, *config.Name, opts)
		},
	}
}

func NewPersistentVolumeClaim(config *acapiv1.PersistentVolumeClaimApplyConfiguration, client typedapiv1.PersistentVolumeClaimInterface) *Resource {
	return &Resource{
		name: *config.Name,
		kind: "pvc",
		applyFunc: func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
			return client.Apply(ctx, config, opts)
		},
		getFunc: func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
			return client.Get(ctx, *config.Name, opts)
		},
	}
}
