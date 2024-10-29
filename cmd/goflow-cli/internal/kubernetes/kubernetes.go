package kubernetes

import (
	"context"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/resource"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type kubeConfigBuilder interface {
	GetKubeConfigPath() (string, error)
	BuildConfig(clusterURL, kubeConfigPath string) (*rest.Config, error)
}

type kubeClientBuilder interface {
	NewForConfig(config *rest.Config) (*kubernetes.Clientset, error)
}

type applyConfiguration interface {
	*acapiv1.NamespaceApplyConfiguration |
		*acappsv1.DeploymentApplyConfiguration |
		*acapiv1.ServiceApplyConfiguration |
		*acapiv1.PersistentVolumeApplyConfiguration |
		*acapiv1.PersistentVolumeClaimApplyConfiguration
}

type resourceApplier[C applyConfiguration] interface {
	Apply(
		ctx context.Context,
		client resource.ApplyWatchable[C, any],
		config C,
		namespace string,
		logger log.Logger,
		waiter resource.EventWaiter,
	) error
}

type KubeClient struct {
	ctx               context.Context
	clientset         kubernetes.Interface
	namespace         string
	logger            log.Logger
	waiter            resource.EventWaiter
	namespaceApplier  resourceApplier[*acapiv1.NamespaceApplyConfiguration]
	deploymentApplier resourceApplier[*acappsv1.DeploymentApplyConfiguration]
	serviceApplier    resourceApplier[*acapiv1.ServiceApplyConfiguration]
	pvApplier         resourceApplier[*acapiv1.PersistentVolumeApplyConfiguration]
	pvcApplier        resourceApplier[*acapiv1.PersistentVolumeClaimApplyConfiguration]
}

func New(
	clusterURL string,
	namespace string,
	namespaceApplier resourceApplier[*acapiv1.NamespaceApplyConfiguration],
	deploymentApplier resourceApplier[*acappsv1.DeploymentApplyConfiguration],
	serviceApplier resourceApplier[*acapiv1.ServiceApplyConfiguration],
	pvApplier resourceApplier[*acapiv1.PersistentVolumeApplyConfiguration],
	pvcApplier resourceApplier[*acapiv1.PersistentVolumeClaimApplyConfiguration],
	opts ...Option,
) (*KubeClient, error) {
	options := defaultOptions()

	for _, o := range opts {
		o.apply(&options)
	}

	kubeConfigPath, err := options.configBuilder.GetKubeConfigPath()
	if err != nil {
		return nil, err
	}

	kubeConfig, err := options.configBuilder.BuildConfig(clusterURL, kubeConfigPath)
	if err != nil {
		return nil, err
	}

	clientset, err := options.kubeClientBuilder.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	eventWaiter := resource.NewWaiter(ctx, options.logger)

	client := &KubeClient{
		ctx:               ctx,
		clientset:         clientset,
		namespace:         namespace,
		logger:            options.logger,
		waiter:            eventWaiter,
		namespaceApplier:  namespaceApplier,
		deploymentApplier: deploymentApplier,
		serviceApplier:    serviceApplier,
		pvApplier:         pvApplier,
		pvcApplier:        pvcApplier,
	}

	return client, nil
}

func (k *KubeClient) ApplyDeployment(deploymentConfig *acappsv1.DeploymentApplyConfiguration) error {
	deployments := k.clientset.AppsV1().Deployments(k.namespace)
	return k.deploymentApplier.Apply(k.ctx, deployments, deploymentConfig, k.namespace, k.logger, k.waiter)
}

func (k *KubeClient) ApplyService(serviceConfig *acapiv1.ServiceApplyConfiguration) error {
	return k.serviceApplier.Apply(k.ctx, serviceConfig, k.namespace, k.logger, k.waiter)
}

func (k *KubeClient) ApplyPV(pvConfig *acapiv1.PersistentVolumeApplyConfiguration) error {
	return k.pvApplier.Apply(k.ctx, pvConfig, k.namespace, k.logger, k.waiter)
}

func (k *KubeClient) ApplyPVC(pvcConfig *acapiv1.PersistentVolumeClaimApplyConfiguration) error {
	return k.pvcApplier.Apply(k.ctx, pvcConfig, k.namespace, k.logger, k.waiter)
}

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
