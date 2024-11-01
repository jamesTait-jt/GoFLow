package resource

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"

	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
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
	Apply(ctx context.Context, namespace *acapiv1.NamespaceApplyConfiguration, opts metav1.ApplyOptions) (*apiv1.Namespace, error)
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

type deploymentInterface interface {
	Apply(ctx context.Context, deployment *acappsv1.DeploymentApplyConfiguration, opts metav1.ApplyOptions) (*appsv1.Deployment, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*appsv1.Deployment, error)
}

func NewDeployment(config *acappsv1.DeploymentApplyConfiguration, client deploymentInterface) *Resource {
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

type serviceInterface interface {
	Apply(ctx context.Context, service *acapiv1.ServiceApplyConfiguration, opts metav1.ApplyOptions) (*apiv1.Service, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiv1.Service, error)
}

func NewService(config *acapiv1.ServiceApplyConfiguration, client serviceInterface) *Resource {
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

type persistentVolumeInterface interface {
	Apply(ctx context.Context, pv *acapiv1.PersistentVolumeApplyConfiguration, opts metav1.ApplyOptions) (*apiv1.PersistentVolume, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiv1.PersistentVolume, error)
}

func NewPersistentVolume(config *acapiv1.PersistentVolumeApplyConfiguration, client persistentVolumeInterface) *Resource {
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

type persistentVolumeClaimInterface interface {
	Apply(ctx context.Context, pvc *acapiv1.PersistentVolumeClaimApplyConfiguration, opts metav1.ApplyOptions) (*apiv1.PersistentVolumeClaim, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiv1.PersistentVolumeClaim, error)
}

func NewPersistentVolumeClaim(config *acapiv1.PersistentVolumeClaimApplyConfiguration, client persistentVolumeClaimInterface) *Resource {
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
