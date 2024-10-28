package run

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes"
	grpcserver "github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/grpc_server"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/redis"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/workerpool"
	"github.com/jamesTait-jt/goflow/pkg/log"
)

// TODO: Accept a deployOpts struct or something
func Deploy(conf *config.Config, logger log.Logger) error {
	return deployKubernetes(conf, logger)
}

func deployKubernetes(conf *config.Config, logger log.Logger) error {
	stopLog := logger.Waiting("Connecting to the Kubernetes cluster")

	kubeClient, err := kubernetes.New(
		conf.Kubernetes.ClusterURL,
		kubernetes.WithLogger(logger),
	)
	if err != nil {
		stopLog("Failed connecting to kubernetes cluster", false)

		return err
	}

	stopLog("Successfully connected to Kubernetes cluster", true)

	logger.Info(fmt.Sprintf("Initialising kubernetes namespace '%s'", conf.Kubernetes.Namespace))

	err = kubeClient.CreateNamespaceIfNotExists(conf.Kubernetes.Namespace)
	if err != nil {
		return err
	}

	kubeClient.InitialiseClients()

	logger.Info("Deploying message broker")

	if err = kubeClient.ApplyDeployment(redis.Deployment(conf)); err != nil {
		return err
	}

	if err = kubeClient.CreateOrUpdateService(redis.Service(conf)); err != nil {
		return err
	}

	logger.Info("Deploying goflow gRPC server")

	if err = kubeClient.ApplyDeployment(grpcserver.Deployment(conf)); err != nil {
		return err
	}

	if err = kubeClient.CreateOrUpdateService(grpcserver.Service(conf)); err != nil {
		return err
	}

	logger.Info("Uploading plugins")

	if err = kubeClient.CreatePV(workerpool.HandlersPV(conf)); err != nil {
		return err
	}

	if err = kubeClient.CreatePVC(workerpool.HandlersPVC(conf)); err != nil {
		return err
	}

	logger.Info("Deploying workerpools")

	if err = kubeClient.ApplyDeployment(workerpool.Deployment(conf)); err != nil {
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
