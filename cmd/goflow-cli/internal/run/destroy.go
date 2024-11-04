package run

import (
	"fmt"
	"time"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/resource"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/workerpool"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"k8s.io/apimachinery/pkg/watch"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

var destroyTimeout = 30 * time.Second

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

	// Deleting the namespace will delete all associated resources
	namespace := resource.NewNamespace(acapiv1.Namespace(conf.Kubernetes.Namespace), namespacesClient)
	if err := destroyAndWait(namespace, kubeOperator, logger); err != nil {
		return err
	}

	// Persistent volumes are not scoped to a namespace
	pv := resource.NewPersistentVolume(workerpool.HandlersPV(conf), pvClient)
	if err := destroyAndWait(pv, kubeOperator, logger); err != nil {
		return err
	}

	logger.Success("GoFlow destroyed!")

	return nil
}

type IdentifiableWatchableDeleter interface {
	Name() string
	Kind() string
	kubernetes.Deleter
	kubernetes.Watcher
}

func destroyAndWait(kubeResource IdentifiableWatchableDeleter, kubeOperator *kubernetes.Operator, logger log.Logger) error {
	logger.Info(fmt.Sprintf("Destroying %s '%s'", kubeResource.Kind(), kubeResource.Name()))

	neededDeleting, err := kubeOperator.Delete(kubeResource)
	if err != nil {
		return err
	}

	if !neededDeleting {
		logger.Warn(fmt.Sprintf("couldnt find '%s'", kubeResource.Name()))

		return nil
	}

	logger.Info(fmt.Sprintf("'%s' needs destroying - waiting...", kubeResource.Name()))

	if err := kubeOperator.WaitFor(kubeResource, []watch.EventType{watch.Deleted}, destroyTimeout); err != nil {
		return err
	}

	logger.Success(fmt.Sprintf("'%s' destroyed successfully", kubeResource.Name()))

	return nil
}
