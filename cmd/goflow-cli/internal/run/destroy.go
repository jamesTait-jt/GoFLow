package run

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/workerpool"
)

func Destroy(conf *config.Config) error {
	kubeClient, err := kubernetes.New(conf.Kubernetes.ClusterURL, nil)
	if err != nil {
		return err
	}

	// This will delete the namespace and everything contained within
	err = kubeClient.DestroyNamespace(conf.Kubernetes.Namespace)
	if err != nil {
		return err
	}

	// Persistent volumes are not associated with a namespace so must be delete individually
	err = kubeClient.DestroyPV(workerpool.HandlersPV(conf).Name)

	return err
}
