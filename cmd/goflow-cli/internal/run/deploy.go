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
	logger.Info("Connecting to the Kubernetes cluster")

	kubeClient, err := kubernetes.New(
		conf.Kubernetes.ClusterURL,
		conf.Kubernetes.Namespace,
		kubernetes.WithLogger(logger),
	)
	if err != nil {
		return err
	}

	// logger.Info(fmt.Sprintf("Initialising namespace '%s'", conf.Kubernetes.Namespace))
	//
	// if err = kubeClient.ApplyNamespace(conf.Kubernetes.Namespace); err != nil {
	// 	return err
	// }

	logger.Info("Deploying message broker")

	if err = kubeClient.ApplyDeployment(redis.Deployment(conf)); err != nil {
		return err
	}

	if err = kubeClient.ApplyService(redis.Service(conf)); err != nil {
		return err
	}

	logger.Info("Deploying goflow gRPC server")

	if err = kubeClient.ApplyDeployment(grpcserver.Deployment(conf)); err != nil {
		return err
	}

	if err = kubeClient.ApplyService(grpcserver.Service(conf)); err != nil {
		return err
	}

	logger.Info("Uploading plugins")

	if err = kubeClient.ApplyPV(workerpool.HandlersPV(conf)); err != nil {
		return err
	}

	if err = kubeClient.ApplyPVC(workerpool.HandlersPVC(conf)); err != nil {
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
