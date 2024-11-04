package run

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/grpcserver"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/redis"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/resource"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/workerpool"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"k8s.io/apimachinery/pkg/watch"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

// TODO: Accept a deployOpts struct or something
func Deploy(conf *config.Config, logger log.Logger) error {
	return deployKubernetes(conf, logger)
}

func deployKubernetes(
	conf *config.Config,
	logger log.Logger,
) error {
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
	deploymentsClient := clientset.AppsV1().Deployments(conf.Kubernetes.Namespace)
	servicesClient := clientset.CoreV1().Services(conf.Kubernetes.Namespace)
	pvClient := clientset.CoreV1().PersistentVolumes()
	pvcClient := clientset.CoreV1().PersistentVolumeClaims(conf.Kubernetes.Namespace)

	namespace := resource.NewNamespace(acapiv1.Namespace(conf.Kubernetes.Namespace), namespacesClient)
	if err := applyAndWait(namespace, kubeOperator, logger); err != nil {
		return err
	}

	if err := deployMessageBroker(conf, kubeOperator, deploymentsClient, servicesClient, logger); err != nil {
		return err
	}

	if err := deployGRPCServer(conf, kubeOperator, deploymentsClient, servicesClient, logger); err != nil {
		return err
	}

	if err := deployWorkerpools(conf, kubeOperator, deploymentsClient, pvClient, pvcClient, logger); err != nil {
		return err
	}

	logger.Success(
		fmt.Sprintf(
			"GoFlow deployed! Use `kubectl get pods -n %s` to see the status of the application",
			conf.Kubernetes.Namespace,
		),
	)

	return nil
}

func deployMessageBroker(
	conf *config.Config,
	kubeOperator *kubernetes.Operator,
	deploymentsClient resource.DeploymentInterface,
	servicesClient resource.ServicesInterface,
	logger log.Logger,
) error {
	deployment := resource.NewDeployment(redis.Deployment(conf), deploymentsClient)
	if err := applyAndWait(deployment, kubeOperator, logger); err != nil {
		return err
	}

	service := resource.NewService(redis.Service(conf), servicesClient)

	return applyAndWait(service, kubeOperator, logger)
}

func deployGRPCServer(
	conf *config.Config,
	kubeOperator *kubernetes.Operator,
	deploymentsClient resource.DeploymentInterface,
	servicesClient resource.ServicesInterface,
	logger log.Logger,
) error {
	deployment := resource.NewDeployment(grpcserver.Deployment(conf), deploymentsClient)
	if err := applyAndWait(deployment, kubeOperator, logger); err != nil {
		return err
	}

	service := resource.NewService(grpcserver.Service(conf), servicesClient)

	return applyAndWait(service, kubeOperator, logger)
}

func deployWorkerpools(
	conf *config.Config,
	kubeOperator *kubernetes.Operator,
	deploymentsClient resource.DeploymentInterface,
	pvClient resource.PersistentVolumeInterface,
	pvcClient resource.PersistentVolumeClaimInterface,
	logger log.Logger,
) error {
	pv := resource.NewPersistentVolume(workerpool.HandlersPV(conf), pvClient)
	if err := applyAndWait(pv, kubeOperator, logger); err != nil {
		return err
	}

	pvc := resource.NewPersistentVolumeClaim(workerpool.HandlersPVC(conf), pvcClient)
	if err := applyAndWait(pvc, kubeOperator, logger); err != nil {
		return err
	}

	deployment := resource.NewDeployment(workerpool.Deployment(conf), deploymentsClient)

	return applyAndWait(deployment, kubeOperator, logger)
}

type IdentifiableWatchableApplyGetter interface {
	Name() string
	Kind() string
	kubernetes.ApplyGetter
	kubernetes.Watcher
}

func applyAndWait(kubeResource IdentifiableWatchableApplyGetter, kubeOperator *kubernetes.Operator, logger log.Logger) error {
	logger.Info(fmt.Sprintf("Deploying %s '%s'", kubeResource.Kind(), kubeResource.Name()))

	neededModification, err := kubeOperator.Apply(kubeResource)
	if err != nil {
		return err
	}

	if !neededModification {
		logger.Success(fmt.Sprintf("'%s' deployed successfully", kubeResource.Name()))

		return nil
	}

	logger.Info(fmt.Sprintf("'%s' needs modification - applying changes", kubeResource.Name()))

	if err := kubeOperator.WaitFor(kubeResource, []watch.EventType{watch.Added, watch.Modified}); err != nil {
		return err
	}

	logger.Success(fmt.Sprintf("'%s' deployed successfully", kubeResource.Name()))

	return nil
}
