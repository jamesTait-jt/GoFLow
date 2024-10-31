package kubernetes

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/resource"
	"github.com/jamesTait-jt/goflow/pkg/log"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type kubeConfigBuilder interface {
	GetKubeConfigPath() (string, error)
	BuildConfig(clusterURL, kubeConfigPath string) (*rest.Config, error)
}

type kubeClientBuilder interface {
	NewForConfig(config *rest.Config) (*kubernetes.Clientset, error)
}

type resourceApplier[A resource.Appliable] interface {
	Apply(
		ctx context.Context,
		toApply A,
	) (bool, error)
}

type KubeClient struct {
	ctx    context.Context
	logger log.Logger
	// waiter            resource.EventWaiter
	speccer           resource.Speccer
	namespaceApplier  resourceApplier[*acapiv1.NamespaceApplyConfiguration]
	deploymentApplier resourceApplier[*acappsv1.DeploymentApplyConfiguration]
	serviceApplier    resourceApplier[*acapiv1.ServiceApplyConfiguration]
	pvApplier         resourceApplier[*acapiv1.PersistentVolumeApplyConfiguration]
	pvcApplier        resourceApplier[*acapiv1.PersistentVolumeClaimApplyConfiguration]
}

func New(
	namespaceApplier resourceApplier[*acapiv1.NamespaceApplyConfiguration],
	deploymentApplier resourceApplier[*acappsv1.DeploymentApplyConfiguration],
	serviceApplier resourceApplier[*acapiv1.ServiceApplyConfiguration],
	// pvApplier resourceApplier[*acapiv1.PersistentVolumeApplyConfiguration],
	// pvcApplier resourceApplier[*acapiv1.PersistentVolumeClaimApplyConfiguration],
	opts ...Option,
) (*KubeClient, error) {
	options := defaultOptions()

	for _, o := range opts {
		o.apply(&options)
	}

	// kubeConfigPath, err := options.configBuilder.GetKubeConfigPath()
	// if err != nil {
	// 	return nil, err
	// }

	// kubeConfig, err := options.configBuilder.BuildConfig(clusterURL, kubeConfigPath)
	// if err != nil {
	// 	return nil, err
	// }

	// clientset, err := options.kubeClientBuilder.NewForConfig(kubeConfig)
	// if err != nil {
	// 	return nil, err
	// }

	ctx := context.Background()

	// eventWaiter := resource.NewWaiter(ctx, options.logger)

	client := &KubeClient{
		ctx:    ctx,
		logger: options.logger,
		// waiter:           eventWaiter,
		speccer:           &resource.ObjectSpeccer{},
		namespaceApplier:  namespaceApplier,
		deploymentApplier: deploymentApplier,
		serviceApplier:    serviceApplier,
		// pvApplier:         pvApplier,
		// pvcApplier:        pvcApplier,
	}

	return client, nil
}

type kubeResource interface {
	Apply(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error)
	Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error)
}

func (k *KubeClient) Apply(r kubeResource) (bool, error) {
	currDeployment, err := r.Get(
		k.ctx,
		metav1.GetOptions{},
	)

	if err != nil && !k8serr.IsNotFound(err) {
		return false, err
	}

	proposedDeployment, err := r.Apply(
		k.ctx,
		metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}},
	)

	currSpec, err := k.speccer.Spec(currDeployment)
	if err != nil {
		return false, err
	}

	proposedSpec, err := k.speccer.Spec(proposedDeployment)
	if err != nil {
		return false, err
	}

	// new deployment is the same as the old deployment
	if reflect.DeepEqual(currSpec, proposedSpec) {
		return false, nil
	}

	_, err = r.Apply(
		k.ctx,
		metav1.ApplyOptions{FieldManager: "goflow-cli"},
	)

	if err != nil {
		return false, err
	}

	return true, err
}

func (k *KubeClient) ApplyNamespace(namespace *acapiv1.NamespaceApplyConfiguration) error {
	neededUpdate, err := k.namespaceApplier.Apply(k.ctx, namespace)

	if err != nil {
		return err
	}

	k.logger.Info(fmt.Sprintf("%s", neededUpdate))

	return nil
}

func (k *KubeClient) ApplyDeployment(deployment *acappsv1.DeploymentApplyConfiguration) error {
	neededUpdate, err := k.deploymentApplier.Apply(k.ctx, deployment)

	if err != nil {
		return err
	}

	k.logger.Info(fmt.Sprintf("%s", neededUpdate))

	return nil
}

func (k *KubeClient) ApplyService(service *acapiv1.ServiceApplyConfiguration) error {
	neededUpdate, err := k.serviceApplier.Apply(k.ctx, service)

	if err != nil {
		return err
	}

	k.logger.Info(fmt.Sprintf("%s", neededUpdate))

	return nil
}

// func (k *KubeClient) ApplyPV(pvConfig *acapiv1.PersistentVolumeApplyConfiguration) error {
// 	return k.pvApplier.Apply(k.ctx, pvConfig, k.namespace, k.logger, k.waiter)
// }

// func (k *KubeClient) ApplyPVC(pvcConfig *acapiv1.PersistentVolumeClaimApplyConfiguration) error {
// 	return k.pvcApplier.Apply(k.ctx, pvcConfig, k.namespace, k.logger, k.waiter)
// }

// func (k *KubeClient) DestroyNamespace(namespace string) error {
// 	namespacesClient := k.client.CoreV1().Namespaces()

// 	err := namespacesClient.Delete(k.ctx, namespace, metav1.DeleteOptions{})
// 	if err != nil {
// 		return err
// 	}

// 	k.namespace = namespace

// 	return k.waiter.WaitFor(namespace, "", []watch.EventType{watch.Deleted}, namespacesClient)
// }

// func (k *KubeClient) CreateOrUpdateService(service *apiv1.Service) error {
// 	servicesClient := k.client.CoreV1().Services(k.namespace)

// 	_, err := servicesClient.Get(k.ctx, service.Name, metav1.GetOptions{})

// 	// Service already exists
// 	if err == nil {
// 		if _, err = servicesClient.Update(k.ctx, service, metav1.UpdateOptions{}); err != nil {
// 			return err
// 		}

// 		return nil
// 	}

// 	if !k8serr.IsNotFound(err) {
// 		return err
// 	}

// 	_, err = servicesClient.Create(k.ctx, service, metav1.CreateOptions{})
// 	if err != nil {
// 		return err
// 	}

// 	return k.waiter.WaitFor(service.Name, k.namespace, []watch.EventType{watch.Added}, servicesClient)
// }

// func (k *KubeClient) CreatePV(pv *apiv1.PersistentVolume) error {
// 	pvClient := k.client.CoreV1().PersistentVolumes()

// 	_, err := pvClient.Create(k.ctx, pv, metav1.CreateOptions{})
// 	if err != nil {
// 		return err
// 	}

// 	return k.waiter.WaitFor(pv.Name, "", []watch.EventType{watch.Added}, pvClient)
// }

// func (k *KubeClient) DestroyPV(pvName string) error {
// 	pvClient := k.client.CoreV1().PersistentVolumes()

// 	err := pvClient.Delete(k.ctx, pvName, metav1.DeleteOptions{})
// 	if err != nil {
// 		return err
// 	}

// 	return k.waiter.WaitFor(pvName, "", []watch.EventType{watch.Deleted}, pvClient)
// }

// func (k *KubeClient) CreatePVC(pvc *apiv1.PersistentVolumeClaim) error {
// 	pvcClient := k.client.CoreV1().PersistentVolumeClaims(k.namespace)

// 	_, err := pvcClient.Create(k.ctx, pvc, metav1.CreateOptions{})
// 	if err != nil {
// 		return err
// 	}

// 	return k.waiter.WaitFor(pvc.Name, k.namespace, []watch.EventType{watch.Added}, pvcClient)
// }
