package resource

import (
	"context"
	"fmt"

	"github.com/jamesTait-jt/goflow/pkg/log"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type ApplierV2 struct {}



type hasName interface {
	GetName() *string
}

type ApplyWatchable[C hasName, R any] interface {
	Apply(
		ctx context.Context,
		configuration C,
		opts metav1.ApplyOptions,
	) (result *R, err error)
	Watchable
}

type EventWaiter interface {
	WaitFor(resourceName, namespace string, eventTypes []watch.EventType, client Watchable) error
}

type Applier[C hasName, R any] struct {
	client ApplyWatchable[C, R]
}

func NewNamespaceApplier(
	clientset kubernetes.Interface,
) *Applier[*acapiv1.NamespaceApplyConfiguration, apiv1.Namespace] {
	return &Applier[*acapiv1.NamespaceApplyConfiguration, apiv1.Namespace]{
		client: clientset.CoreV1().Namespaces(),
	}
}

func NewDeploymentApplier(
	clientset kubernetes.Interface,
	namespace string,
) *Applier[*acappsv1.DeploymentApplyConfiguration, appsv1.Deployment] {
	return &Applier[*acappsv1.DeploymentApplyConfiguration, appsv1.Deployment]{
		client: clientset.AppsV1().Deployments(namespace),
	}
}

func NewServiceApplier(
	clientset kubernetes.Interface,
	namespace string,
) *Applier[*acapiv1.ServiceApplyConfiguration, apiv1.Service] {
	return &Applier[*acapiv1.ServiceApplyConfiguration, apiv1.Service]{
		client: clientset.CoreV1().Services(namespace),
	}
}

func (a *Applier[C, R]) Apply(
	ctx context.Context, config C, namespace string, logger log.Logger, waiter EventWaiter,
) error {
	changes, err := a.client.Apply(
		ctx, config, metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}},
	)

	if err != nil {
		return err
	}

	// no changes required
	if changes == nil {
		logger.Info(fmt.Sprintf("No changes required for '%s'", *config.GetName()))

		return nil
	}

	_, err = a.client.Apply(
		ctx, config, metav1.ApplyOptions{FieldManager: "goflow-cli"},
	)

	if err != nil {
		return err
	}

	return waiter.WaitFor(
		*config.GetName(),
		namespace,
		[]watch.EventType{watch.Added, watch.Modified},
		a.client,
	)
}
