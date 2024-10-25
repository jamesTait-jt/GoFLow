package run

import (
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/docker"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes"
	grpcserver "github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/grpc_server"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/redis"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/workerpool"
)

// TODO: Make this accept a deploytOpts struct or something
func Deploy(conf *config.Config) error {
	return deployKubernetes(conf)
}

func deployKubernetes(conf *config.Config) error {
	kubeClient, err := kubernetes.New(conf.Kubernetes.ClusterURL, conf.Kubernetes.Namespace)
	if err != nil {
		return err
	}

	fmt.Println("Kubernetes client created successfully!")

	fmt.Println("Starting Redis...")

	if err := kubeClient.CreateOrReplaceDeployment(redis.Deployment(conf)); err != nil {
		return err
	}

	if err = kubeClient.CreateOrReplaceService(redis.Service(conf)); err != nil {
		return err
	}

	fmt.Println("Starting goflow gRPC service...")

	if err = kubeClient.CreateOrReplaceDeployment(grpcserver.Deployment(conf)); err != nil {
		return err
	}

	if err = kubeClient.CreateOrReplaceService(grpcserver.Service(conf)); err != nil {
		return err
	}

	fmt.Println("Uploading plugins...")

	if err = kubeClient.CreateOrReplacePV(workerpool.HandlersPV(conf)); err != nil {
		return err
	}

	if err = kubeClient.CreateOrReplacePVC(workerpool.HandlersPVC(conf)); err != nil {
		return err
	}

	fmt.Println("Starting workerpool...")

	if err = kubeClient.CreateOrReplaceDeployment(workerpool.Deployment(conf)); err != nil {
		return err
	}

	return nil
}
