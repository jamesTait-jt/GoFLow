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

// TODO: Make this accept a deploytOpts struct or something
func Deploy(conf *config.Config, logger log.Logger) error {
	return deployKubernetes(conf, logger)
}

func deployKubernetes(conf *config.Config, logger log.Logger) error {
	stopLog := logger.Waiting("Connecting to the Kubernetes cluster")

	kubeClient, err := kubernetes.New(conf.Kubernetes.ClusterURL, logger)
	if err != nil {
		stopLog("Failed connecting to kubernetes cluster", false)

		return err
	}

	stopLog("Successfully connected to Kubernetes cluster", true)

	logger.Info("Initialising kubernetes namespace")

	err = kubeClient.CreateNamespaceIfNotExists(conf.Kubernetes.Namespace)
	if err != nil {
		return err
	}

	stopLog = logger.Waiting("Deploying message broker")

	if err := kubeClient.CreateOrUpdateDeployment(redis.Deployment(conf)); err != nil {
		stopLog("Failed deploying message broker", false)
		return err
	}

	if err = kubeClient.CreateOrUpdateService(redis.Service(conf)); err != nil {
		stopLog("Failed deploying message broker", false)
		return err
	}

	stopLog("Successfully deployed message broker", true)

	stopLog = logger.Waiting("Deploying goflow gRPC server")

	if err = kubeClient.CreateOrUpdateDeployment(grpcserver.Deployment(conf)); err != nil {
		stopLog("Failed deploying gRPC server", false)
		return err
	}

	if err = kubeClient.CreateOrUpdateService(grpcserver.Service(conf)); err != nil {
		stopLog("Failed deploying gRPC server", false)
		return err
	}

	stopLog("Successfully deployed gRPC server", true)

	stopLog = logger.Waiting("Uploading plugins")

	if err = kubeClient.CreatePV(workerpool.HandlersPV(conf)); err != nil {
		stopLog("Failed uploading plugins", false)
		return err
	}

	if err = kubeClient.CreatePVC(workerpool.HandlersPVC(conf)); err != nil {
		stopLog("Failed uploading plugins", false)
		return err
	}

	stopLog("Successfully uploaded plugins", true)

	stopLog = logger.Waiting("Deploying workerpools")

	if err = kubeClient.CreateOrUpdateDeployment(workerpool.Deployment(conf)); err != nil {
		stopLog("Failed deploying workerpools", false)
		return err
	}

	stopLog("Susccessfully deployed workerpools", true)

	logger.Success(
		fmt.Sprintf(
			"GoFlow deployed! Use `kubectl get pods -n %s` to see the status of the application",
			conf.Kubernetes.Namespace,
		),
	)

	return nil
}
