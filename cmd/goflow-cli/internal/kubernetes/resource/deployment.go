package resource

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

type Deployment struct {
	config *acappsv1.DeploymentApplyConfiguration
	client typedappsv1.DeploymentInterface
}

func NewDeployment(config *acappsv1.DeploymentApplyConfiguration, client typedappsv1.DeploymentInterface) *Deployment {
	return &Deployment{
		config: config,
		client: client,
	}
}

func (d *Deployment) Name() string {
	return *d.config.Name
}

func (d *Deployment) Kind() string {
	return "deployment"
}

func (d *Deployment) Apply(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
	return d.client.Apply(ctx, d.config, opts)
}

func (d *Deployment) Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
	return d.client.Get(ctx, d.Name(), opts)
}
