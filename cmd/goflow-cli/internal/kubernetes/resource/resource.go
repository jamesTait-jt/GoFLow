package resource

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"

	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"

	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type Resource struct {
	name       string
	namespace  string
	kind       string
	applyFunc  func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error)
	deleteFunc func(ctx context.Context, opts metav1.DeleteOptions) error
	getFunc    func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error)
	watchFunc  func(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
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

func (r *Resource) Delete(ctx context.Context, opts metav1.DeleteOptions) error {
	return r.deleteFunc(ctx, opts)
}

func (r *Resource) Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
	return r.getFunc(ctx, opts)
}

func (r *Resource) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var fieldSelector string

	if opts.FieldSelector != "" {
		return r.watchFunc(ctx, opts)
	}

	if r.namespace != "" {
		fieldSelector = fmt.Sprintf("metadata.name=%s,metadata.namespace=%s", r.name, r.namespace)
	} else {
		fieldSelector = fmt.Sprintf("metadata.name=%s", r.name)
	}

	opts.FieldSelector = fieldSelector

	return r.watchFunc(ctx, opts)
}

type baseInterface interface {
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type NamespaceInterface interface {
	baseInterface
	Apply(ctx context.Context, namespace *acapiv1.NamespaceApplyConfiguration, opts metav1.ApplyOptions) (*apiv1.Namespace, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiv1.Namespace, error)
}

func NewNamespace(config *acapiv1.NamespaceApplyConfiguration, client NamespaceInterface) *Resource {
	name := *config.Name

	return &Resource{
		name: name,
		kind: "namespace",
		applyFunc: func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
			return client.Apply(ctx, config, opts)
		},
		deleteFunc: func(ctx context.Context, opts metav1.DeleteOptions) error {
			return client.Delete(ctx, name, opts)
		},
		getFunc: func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
			return client.Get(ctx, name, opts)
		},
		watchFunc: func(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
			return client.Watch(ctx, opts)
		},
	}
}

type DeploymentInterface interface {
	baseInterface
	Apply(ctx context.Context, deployment *acappsv1.DeploymentApplyConfiguration, opts metav1.ApplyOptions) (*appsv1.Deployment, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*appsv1.Deployment, error)
}

func NewDeployment(config *acappsv1.DeploymentApplyConfiguration, client DeploymentInterface) *Resource {
	name := *config.Name

	return &Resource{
		name: name,
		kind: "deployment",
		applyFunc: func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
			return client.Apply(ctx, config, opts)
		},
		deleteFunc: func(ctx context.Context, opts metav1.DeleteOptions) error {
			return client.Delete(ctx, name, opts)
		},
		getFunc: func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
			return client.Get(ctx, name, opts)
		},
		watchFunc: func(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
			return client.Watch(ctx, opts)
		},
	}
}

type ServicesInterface interface {
	baseInterface
	Apply(ctx context.Context, service *acapiv1.ServiceApplyConfiguration, opts metav1.ApplyOptions) (*apiv1.Service, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiv1.Service, error)
}

func NewService(config *acapiv1.ServiceApplyConfiguration, client ServicesInterface) *Resource {
	name := *config.Name

	return &Resource{
		name: name,
		kind: "service",
		applyFunc: func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
			return client.Apply(ctx, config, opts)
		},
		deleteFunc: func(ctx context.Context, opts metav1.DeleteOptions) error {
			return client.Delete(ctx, name, opts)
		},
		getFunc: func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
			return client.Get(ctx, name, opts)
		},
		watchFunc: func(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
			return client.Watch(ctx, opts)
		},
	}
}

type PersistentVolumeInterface interface {
	baseInterface
	Apply(ctx context.Context, pv *acapiv1.PersistentVolumeApplyConfiguration, opts metav1.ApplyOptions) (*apiv1.PersistentVolume, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiv1.PersistentVolume, error)
}

func NewPersistentVolume(config *acapiv1.PersistentVolumeApplyConfiguration, client PersistentVolumeInterface) *Resource {
	name := *config.Name

	return &Resource{
		name: name,
		kind: "pv",
		applyFunc: func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
			return client.Apply(ctx, config, opts)
		},
		deleteFunc: func(ctx context.Context, opts metav1.DeleteOptions) error {
			return client.Delete(ctx, name, opts)
		},
		getFunc: func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
			return client.Get(ctx, name, opts)
		},
		watchFunc: func(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
			return client.Watch(ctx, opts)
		},
	}
}

type PersistentVolumeClaimInterface interface {
	baseInterface
	Apply(ctx context.Context, pvc *acapiv1.PersistentVolumeClaimApplyConfiguration, opts metav1.ApplyOptions) (*apiv1.PersistentVolumeClaim, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiv1.PersistentVolumeClaim, error)
}

func NewPersistentVolumeClaim(config *acapiv1.PersistentVolumeClaimApplyConfiguration, client PersistentVolumeClaimInterface) *Resource {
	name := *config.Name

	return &Resource{
		name: name,
		kind: "pvc",
		applyFunc: func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
			return client.Apply(ctx, config, opts)
		},
		deleteFunc: func(ctx context.Context, opts metav1.DeleteOptions) error {
			return client.Delete(ctx, name, opts)
		},
		getFunc: func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
			return client.Get(ctx, name, opts)
		},
		watchFunc: func(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
			return client.Watch(ctx, opts)
		},
	}
}
