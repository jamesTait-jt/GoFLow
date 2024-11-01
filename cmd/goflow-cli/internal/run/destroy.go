package run

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/resource"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/workerpool"
	"github.com/jamesTait-jt/goflow/pkg/log"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

func Destroy(conf *config.Config, logger log.Logger) error {
	logger.Info("Connecting to the Kubernetes cluster")

	clientset, err := kubernetes.NewClientset(conf.Kubernetes.ClusterURL)
	if err != nil {
		return err
	}

	kubeOperator, err := kubernetes.NewOperator(kubernetes.WithLogger(logger))
	if err != nil {
		return err
	}

	namespacesClient := clientset.CoreV1().Namespaces()
	pvClient := clientset.CoreV1().PersistentVolumes()

	resources := []kubernetes.Resource{
		// Deleting the namespace will delete all associated resources
		resource.NewNamespace(
			acapiv1.Namespace(conf.Kubernetes.Namespace),
			namespacesClient,
		),
		// Persistent volumes are not scoped to a namespace
		resource.NewPersistentVolume(
			workerpool.HandlersPV(conf),
			pvClient,
		),
	}

	for i := 0; i < len(resources); i++ {
		r := resources[i]
		logger.Info(fmt.Sprintf("Destroying %s '%s'", r.Kind(), r.Name()))

		neededDeletion, err := kubeOperator.Delete(r)
		if err != nil {
			return err
		}

		if neededDeletion {
			logger.Info(fmt.Sprintf("'%s' needs destroying - waiting...", r.Name()))
		}

		logger.Success(fmt.Sprintf("'%s' destroyed successfully", r.Name()))
	}

	logger.Success("GoFlow destroyed!")

	return nil
}
