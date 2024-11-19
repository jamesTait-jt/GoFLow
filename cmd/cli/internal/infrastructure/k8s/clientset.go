package k8s

import (
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/infrastructure/k8s/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type kubeConfigBuilder interface {
	GetKubeConfigPath() (string, error)
	BuildConfig(clusterURL, kubeConfigPath string) (*rest.Config, error)
}

type clientSetBuilder interface {
	NewForConfig(config *rest.Config) (*kubernetes.Clientset, error)
}

type Clients struct {
	clientset kubernetes.Interface
	namespace string
}

func NewClientset(clusterURL, namespace string, opts ...BuildClientsetOption) (*Clients, error) {
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

	return &Clients{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// TODO: Optimise this so they are only called once
func (c *Clients) Namespaces() resource.NamespaceInterface {
	return c.clientset.CoreV1().Namespaces()
}
func (c *Clients) Deployments() resource.DeploymentInterface {
	return c.clientset.AppsV1().Deployments(c.namespace)
}
func (c *Clients) Services() resource.ServiceInterface {
	return c.clientset.CoreV1().Services(c.namespace)
}
func (c *Clients) PersistentVolumes() resource.PersistentVolumeInterface {
	return c.clientset.CoreV1().PersistentVolumes()
}
func (c *Clients) PersistentVolumeClaims() resource.PersistentVolumeClaimInterface {
	return c.clientset.CoreV1().PersistentVolumeClaims(c.namespace)
}
