package kubernetes

import (
	"context"
	"fmt"

	"github.com/jamesTait-jt/goflow/pkg/log"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
)

type kubeConfigBuilder interface {
	GetKubeConfigPath() (string, error)
	BuildConfig(clusterURL, kubeConfigPath string) (*rest.Config, error)
}

type kubeClientBuilder interface {
	NewForConfig(config *rest.Config) (*kubernetes.Clientset, error)
}

type watchable interface {
	Watch(ctx context.Context, options metav1.ListOptions) (watch.Interface, error)
}

type deploymentsClient interface {
	Apply(
		ctx context.Context,
		deployment *acappsv1.DeploymentApplyConfiguration,
		opts metav1.ApplyOptions,
	) (result *appsv1.Deployment, err error)
	watchable
}

type eventWaiter interface {
	WaitFor(resourceName, namespace string, eventTypes []watch.EventType, client watchable) error
}

type KubeClient struct {
	ctx               context.Context
	client            *kubernetes.Clientset
	deploymentsClient deploymentsClient
	namespace         string
	waiter            eventWaiter
	logger            log.Logger
}

func New(clusterURL string, opts ...Option) (*KubeClient, error) {
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
	client := &KubeClient{
		ctx:    ctx,
		client: clientset,
		waiter: NewWaiter(ctx, options.logger),
		logger: options.logger,
	}

	return client, nil
}

func (k *KubeClient) CreateNamespaceIfNotExists(namespace string) error {
	namespacesClient := k.client.CoreV1().Namespaces()

	_, err := namespacesClient.Get(k.ctx, namespace, metav1.GetOptions{})
	if err == nil {
		k.namespace = namespace

		k.logger.Success(fmt.Sprintf("Namespace '%s' already exists; proceeding with deployment.", namespace))

		return nil
	}

	if !k8serr.IsNotFound(err) {
		return err
	}

	namespaceObject := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	_, err = namespacesClient.Create(k.ctx, namespaceObject, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	err = k.waiter.WaitFor(namespace, "", []watch.EventType{watch.Added}, namespacesClient)
	if err != nil {
		return err
	}

	k.namespace = namespace

	return nil
}

func (k *KubeClient) DestroyNamespace(namespace string) error {
	namespacesClient := k.client.CoreV1().Namespaces()

	err := namespacesClient.Delete(k.ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	k.namespace = namespace

	return k.waiter.WaitFor(namespace, "", []watch.EventType{watch.Deleted}, namespacesClient)
}

func (k *KubeClient) InitialiseClients() {
	k.deploymentsClient = k.client.AppsV1().Deployments(k.namespace)
}

func (k *KubeClient) ApplyDeployment(deploymentConfig *acappsv1.DeploymentApplyConfiguration) error {
	deployment, err := k.deploymentsClient.Apply(
		k.ctx, deploymentConfig, metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}},
	)

	if err != nil {
		return err
	}

	// no changes required
	if deployment == nil {
		k.logger.Info(fmt.Sprintf("No changes required for deployment '%s'", *deploymentConfig.Name))

		return nil
	}

	_, err = k.deploymentsClient.Apply(
		k.ctx, deploymentConfig, metav1.ApplyOptions{FieldManager: "goflow-cli"},
	)

	if err != nil {
		return err
	}

	return k.waiter.WaitFor(
		*deploymentConfig.Name,
		k.namespace,
		[]watch.EventType{watch.Added, watch.Modified},
		k.deploymentsClient,
	)
}

func (k *KubeClient) CreateOrUpdateService(service *apiv1.Service) error {
	servicesClient := k.client.CoreV1().Services(k.namespace)

	_, err := servicesClient.Get(k.ctx, service.Name, metav1.GetOptions{})

	// Service already exists
	if err == nil {
		if _, err = servicesClient.Update(k.ctx, service, metav1.UpdateOptions{}); err != nil {
			return err
		}

		return nil
	}

	if !k8serr.IsNotFound(err) {
		return err
	}

	_, err = servicesClient.Create(k.ctx, service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return k.waiter.WaitFor(service.Name, k.namespace, []watch.EventType{watch.Added}, servicesClient)
}

func (k *KubeClient) CreatePV(pv *apiv1.PersistentVolume) error {
	pvClient := k.client.CoreV1().PersistentVolumes()

	_, err := pvClient.Create(k.ctx, pv, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return k.waiter.WaitFor(pv.Name, "", []watch.EventType{watch.Added}, pvClient)
}

func (k *KubeClient) DestroyPV(pvName string) error {
	pvClient := k.client.CoreV1().PersistentVolumes()

	err := pvClient.Delete(k.ctx, pvName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return k.waiter.WaitFor(pvName, "", []watch.EventType{watch.Deleted}, pvClient)
}

func (k *KubeClient) CreatePVC(pvc *apiv1.PersistentVolumeClaim) error {
	pvcClient := k.client.CoreV1().PersistentVolumeClaims(k.namespace)

	_, err := pvcClient.Create(k.ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return k.waiter.WaitFor(pvc.Name, k.namespace, []watch.EventType{watch.Added}, pvcClient)
}
