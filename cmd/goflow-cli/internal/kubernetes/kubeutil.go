package kubernetes

import (
	"errors"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type KubeConfigBuilder struct{}

func (k *KubeConfigBuilder) GetKubeConfigPath() (string, error) {
	home := homedir.HomeDir()
	if home == "" {
		return "", errors.New("could not find .kube/config file in home directory")
	}

	return filepath.Join(home, ".kube", "config"), nil
}

func (k *KubeConfigBuilder) BuildConfig(clusterURL, kubeConfPath string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags(clusterURL, kubeConfPath)
}

type KubeClientBuilder struct{}

func (k *KubeClientBuilder) NewForConfig(config *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(config)
}
