package k8s

import (
	"k8s.io/client-go/kubernetes"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedapiv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type kubeConfigBuilder interface {
	GetKubeConfigPath() (string, error)
	BuildConfig(clusterURL, kubeConfigPath string) (*rest.Config, error)
}

type clientSetBuilder interface {
	NewForConfig(config *rest.Config) (*kubernetes.Clientset, error)
}

type K8sClients struct {
	clientset kubernetes.Interface
	namespace string
}

func NewClientset(clusterURL, namespace string, opts ...BuildClientsetOption) (*K8sClients, error) {
	options := defaultBuildClientsetOptions()

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

	return &K8sClients{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

func (c *K8sClients) Namespaces() typedapiv1.NamespaceInterface {
	return c.clientset.CoreV1().Namespaces()
}
func (c *K8sClients) Deployments() typedappsv1.DeploymentInterface {
	return c.clientset.AppsV1().Deployments(c.namespace)
}
func (c *K8sClients) Services() typedapiv1.ServiceInterface {
	return c.clientset.CoreV1().Services(c.namespace)
}
func (c *K8sClients) PersistentVolumes() typedapiv1.PersistentVolumeInterface {
	return c.clientset.CoreV1().PersistentVolumes()
}
func (c *K8sClients) PersistentVolumeClaims() typedapiv1.PersistentVolumeClaimInterface {
	return c.clientset.CoreV1().PersistentVolumeClaims(c.namespace)
}
