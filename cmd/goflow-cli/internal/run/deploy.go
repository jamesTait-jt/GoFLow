package run

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes"
	grpcserver "github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/grpc_server"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/resource"
	"github.com/jamesTait-jt/goflow/pkg/log"
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

	kubeConfigPath, err := kubernetes.GetKubeConfigPath()
	if err != nil {
		return err
	}

	kubeConfig, err := kubernetes.BuildConfig(conf.Kubernetes.ClusterURL, kubeConfigPath)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	kubeClient, err := kubernetes.New(
		resource.NewNamespaceApplier(clientset),
		resource.NewDeploymentApplier(clientset, conf.Kubernetes.Namespace),
		resource.NewServiceApplier(clientset, conf.Kubernetes.Namespace),
	)
	if err != nil {
		return err
	}

	// logger.Info(fmt.Sprintf("Initialising namespace '%s'", conf.Kubernetes.Namespace))

	// if err = kubeClient.ApplyNamespace(acapiv1.Namespace(conf.Kubernetes.Namespace)); err != nil {
	// 	return err
	// }

	logger.Info("Deploying message broker")

	deploymentsClient := clientset.AppsV1().Deployments(conf.Kubernetes.Namespace)

	grpcServer := grpcserver.NewDeploymentApplier(
		grpcserver.Deployment(conf),
		deploymentsClient,
	)

	if _, err = kubeClient.Apply(grpcServer); err != nil {
		return err
	}

	// if err = kubeClient.ApplyService(redis.Service(conf)); err != nil {
	// 	return err
	// }

	// logger.Info("Deploying goflow gRPC server")

	// if err = kubeClient.ApplyDeployment(grpcserver.Deployment(conf)); err != nil {
	// 	return err
	// }

	// if err = kubeClient.ApplyService(grpcserver.Service(conf)); err != nil {
	// 	return err
	// }

	// logger.Info("Uploading plugins")

	// if err = kubeClient.ApplyPV(workerpool.HandlersPV(conf)); err != nil {
	// 	return err
	// }

	// if err = kubeClient.ApplyPVC(workerpool.HandlersPVC(conf)); err != nil {
	// 	return err
	// }

	// logger.Info("Deploying workerpools")

	// if err = kubeClient.ApplyDeployment(workerpool.Deployment(conf)); err != nil {
	// 	return err
	// }

	// logger.Success(
	// 	fmt.Sprintf(
	// 		"GoFlow deployed! Use `kubectl get pods -n %s` to see the status of the application",
	// 		conf.Kubernetes.Namespace,
	// 	),
	// )

	return nil
}
