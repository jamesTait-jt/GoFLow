package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/jamesTait-jt/goflow/pkg/log"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type KubeClient struct {
	ctx       context.Context
	client    *kubernetes.Clientset
	namespace string
	logger    log.Logger
}

func New(clusterURL string, logger log.Logger) (*KubeClient, error) {
	var kubeConfPath string
	if home := homedir.HomeDir(); home != "" {
		kubeConfPath = filepath.Join(home, ".kube", "config")
	} else {
		return nil, errors.New("could not find .kube/config file in home directory")
	}

	kubeConf, err := clientcmd.BuildConfigFromFlags(clusterURL, kubeConfPath)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		return nil, err
	}

	client := &KubeClient{
		ctx:    context.Background(),
		client: clientset,
		logger: logger,
	}

	return client, nil
}

func (k *KubeClient) CreateOrUpdateDeployment(deployment *appsv1.Deployment) error {
	deploymentsClient := k.client.AppsV1().Deployments(k.namespace)

	_, err := deploymentsClient.Create(k.ctx, deployment, metav1.CreateOptions{})
	if err == nil {
		return k.waitFor(deployment.Name, k.namespace, watch.Added, deploymentsClient)
	}

	if !k8serr.IsAlreadyExists(err) {
		return err
	}

	// fmt.Println("Deployment existed - replacing")

	_, err = deploymentsClient.Update(k.ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil
	}

	// fmt.Println("Deployment replaced successfully!")

	return err
}

func (k *KubeClient) CreateOrUpdateService(service *apiv1.Service) error {
	servicesClient := k.client.CoreV1().Services(k.namespace)

	_, err := servicesClient.Get(k.ctx, service.Name, metav1.GetOptions{})

	if err == nil {
		// fmt.Println("Service existed - replacing")

		if _, err = servicesClient.Update(k.ctx, service, metav1.UpdateOptions{}); err != nil {
			return err
		}

		// fmt.Println("Service replaced successfully!")

		return nil
	}

	if !k8serr.IsNotFound(err) {
		return err
	}

	_, err = servicesClient.Create(k.ctx, service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return k.waitFor(service.Name, k.namespace, watch.Added, servicesClient)
}

func (k *KubeClient) CreatePV(pv *apiv1.PersistentVolume) error {
	pvClient := k.client.CoreV1().PersistentVolumes()

	_, err := pvClient.Create(k.ctx, pv, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return k.waitFor(pv.Name, "", watch.Added, pvClient)
}

func (k *KubeClient) DestroyPV(pvName string) error {
	pvClient := k.client.CoreV1().PersistentVolumes()

	err := pvClient.Delete(k.ctx, pvName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return k.waitFor(pvName, "", watch.Deleted, pvClient)
}

func (k *KubeClient) CreatePVC(pvc *apiv1.PersistentVolumeClaim) error {
	pvcClient := k.client.CoreV1().PersistentVolumeClaims(k.namespace)

	_, err := pvcClient.Create(k.ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return k.waitFor(pvc.Name, k.namespace, watch.Added, pvcClient)
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

	stopLog := k.logger.Waiting(fmt.Sprintf("Creating namespace '%s'", namespace))

	_, err = namespacesClient.Create(k.ctx, namespaceObject, metav1.CreateOptions{})
	if err != nil {
		stopLog("Failed to create namespace", false)

		return err
	}

	err = k.waitFor(namespace, "", watch.Added, namespacesClient)
	if err != nil {
		stopLog("Failed waiting for namespace creation", false)

		return err
	}

	stopLog(fmt.Sprintf("Successfully created namespace '%s'", namespace), true)

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

	return k.waitFor(namespace, "", watch.Deleted, namespacesClient)
}

type watchable interface {
	Watch(ctx context.Context, options metav1.ListOptions) (watch.Interface, error)
}

func (k *KubeClient) waitFor(resourceName, namespace string, eventType watch.EventType, client watchable) error {
	// This won't be populated if we're destroying or creating a namespace
	var fieldSelector string
	if namespace != "" {
		fieldSelector = fmt.Sprintf("metadata.name=%s,metadata.namespace=%s", resourceName, namespace)
	} else {
		fieldSelector = fmt.Sprintf("metadata.name=%s", resourceName)
	}

	watcher, err := client.Watch(k.ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for event := range watcher.ResultChan() {
		if event.Type == eventType {
			return nil
		}
	}

	// TODO: This will never reach as we don't have a timeout
	return nil
}
