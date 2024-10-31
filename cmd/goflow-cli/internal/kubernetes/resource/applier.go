package resource

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Appliable interface {
	GetName() *string
}

type getApplier[C Appliable, R runtime.Object] interface {
	Get(
		ctx context.Context,
		name string,
		opts metav1.GetOptions,
	) (R, error)
	Apply(
		ctx context.Context,
		applyConfig C,
		opts metav1.ApplyOptions,
	) (R, error)
}

type Speccer interface {
	Spec(obj runtime.Object) (any, error)
}

type Applier[C Appliable, R runtime.Object] struct {
	client  getApplier[C, R]
	speccer Speccer
}

func (a *Applier[C, R]) Apply(
	ctx context.Context,
	toApply C,
) (bool, error) {
	currDeployment, err := a.client.Get(
		ctx,
		*toApply.GetName(),
		metav1.GetOptions{},
	)

	if err != nil && !k8serr.IsNotFound(err) {
		return false, err
	}

	proposedDeployment, err := a.client.Apply(
		ctx,
		toApply,
		metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}},
	)

	currSpec, err := a.speccer.Spec(currDeployment)
	if err != nil {
		return false, err
	}

	proposedSpec, err := a.speccer.Spec(proposedDeployment)
	if err != nil {
		return false, err
	}

	// new deployment is the same as the old deployment
	if reflect.DeepEqual(currSpec, proposedSpec) {
		return false, nil
	}

	_, err = a.client.Apply(
		ctx,
		toApply,
		metav1.ApplyOptions{FieldManager: "goflow-cli"},
	)

	if err != nil {
		return false, err
	}

	return true, err
}

func NewNamespaceApplier(
	clientset kubernetes.Interface,
) *Applier[*acapiv1.NamespaceApplyConfiguration, *apiv1.Namespace] {
	return &Applier[*acapiv1.NamespaceApplyConfiguration, *apiv1.Namespace]{
		client:  clientset.CoreV1().Namespaces(),
		speccer: &ObjectSpeccer{},
	}
}

func NewDeploymentApplier(
	clientset kubernetes.Interface,
	namespace string,
) *Applier[*acappsv1.DeploymentApplyConfiguration, *appsv1.Deployment] {
	return &Applier[*acappsv1.DeploymentApplyConfiguration, *appsv1.Deployment]{
		client:  clientset.AppsV1().Deployments(namespace),
		speccer: &ObjectSpeccer{},
	}
}

func NewServiceApplier(
	clientset kubernetes.Interface,
	namespace string,
) *Applier[*acapiv1.ServiceApplyConfiguration, *apiv1.Service] {
	return &Applier[*acapiv1.ServiceApplyConfiguration, *apiv1.Service]{
		client:  clientset.CoreV1().Services(namespace),
		speccer: &ObjectSpeccer{},
	}
}
