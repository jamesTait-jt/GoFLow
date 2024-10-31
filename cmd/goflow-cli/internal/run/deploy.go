package run

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes"
	grpcserver "github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/grpc_server"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/redis"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/resource"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/workerpool"
	"github.com/jamesTait-jt/goflow/pkg/log"
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

	resources := []kubernetes.Resource{
		resource.NewNamespace(
			acapiv1.Namespace(conf.Kubernetes.Namespace),
			namespacesClient,
		),
		resource.NewDeployment(
			redis.Deployment(conf),
			deploymentsClient,
		),
		resource.NewService(
			redis.Service(conf),
			servicesClient,
		),
		resource.NewDeployment(
			grpcserver.Deployment(conf),
			deploymentsClient,
		),
		resource.NewService(
			grpcserver.Service(conf),
			servicesClient,
		),
		resource.NewPersistentVolume(
			workerpool.HandlersPV(conf),
			pvClient,
		),
		resource.NewPersistentVolumeClaim(
			workerpool.HandlersPVC(conf),
			pvcClient,
		),
		resource.NewDeployment(
			workerpool.Deployment(conf),
			deploymentsClient,
		),
	}

	for i := 0; i < len(resources); i++ {
		r := resources[i]
		logger.Info(fmt.Sprintf("Deploying %s '%s'", r.Kind(), r.Name()))

		neededModification, err := kubeOperator.Apply(r)
		if err != nil {
			return err
		}

		if neededModification {
			logger.Info(fmt.Sprintf("'%s' needs modification - waiting...", r.Name()))
		}

		logger.Success(fmt.Sprintf("'%s' deployed successfully", r.Name()))
	}

	logger.Success(
		fmt.Sprintf(
			"GoFlow deployed! Use `kubectl get pods -n %s` to see the status of the application",
			conf.Kubernetes.Namespace,
		),
	)

	return nil
}
