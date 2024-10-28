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
	logger.Info("🟢 Connecting to the Kubernetes cluster...")

	kubeClient, err := kubernetes.New(conf.Kubernetes.ClusterURL, logger)
	if err != nil {
		return err
	}

	logger.Info("✅ Successfully connected to Kubernetes cluster")

	logger.Info("🏗️ Initialising kubernetes namespace...")

	err = kubeClient.CreateNamespaceIfNotExists(conf.Kubernetes.Namespace)
	if err != nil {
		return err
	}

	fmt.Printf("Created namespace '%s'\n", conf.Kubernetes.Namespace)

	fmt.Println("Starting Redis...")

	if err := kubeClient.CreateOrUpdateDeployment(redis.Deployment(conf)); err != nil {
		return err
	}

	if err = kubeClient.CreateOrUpdateService(redis.Service(conf)); err != nil {
		return err
	}

	fmt.Println("Starting goflow gRPC service...")

	if err = kubeClient.CreateOrUpdateDeployment(grpcserver.Deployment(conf)); err != nil {
		return err
	}

	if err = kubeClient.CreateOrUpdateService(grpcserver.Service(conf)); err != nil {
		return err
	}

	fmt.Println("Uploading plugins...")

	if err = kubeClient.CreatePV(workerpool.HandlersPV(conf)); err != nil {
		return err
	}

	if err = kubeClient.CreatePVC(workerpool.HandlersPVC(conf)); err != nil {
		return err
	}

	fmt.Println("Starting workerpool...")

	if err = kubeClient.CreateOrUpdateDeployment(workerpool.Deployment(conf)); err != nil {
		return err
	}

	fmt.Printf(
		"GoFlow deployed! Use `kubectl get pods -n %s` to see the status of the application\n",
		conf.Kubernetes.Namespace,
	)

	return nil
}
