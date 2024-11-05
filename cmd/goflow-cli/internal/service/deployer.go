package service

import (
	"sync"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s"
	"github.com/jamesTait-jt/goflow/pkg/log"
)

type deploymentClient interface {
	Deploy() error
	Destroy() error
}

type DeploymentService struct {
	conf       *config.Config
	logger     log.Logger
	client     deploymentClient
	initClient sync.Once
}

func NewDeploymentService(conf *config.Config, logger log.Logger) *DeploymentService {
	return &DeploymentService{
		conf:   conf,
		logger: logger,
	}
}

// initialiseClient will lazily initialise the client
func (d *DeploymentService) initializeClient(conf *config.Config, logger log.Logger) error {
	var initErr error

	d.initClient.Do(func() {
		manager, err := d.initialiseK8sClient(conf, logger)
		if err != nil {
			initErr = err
			return
		}

		d.client = manager
	})

	return initErr
}

func (d *DeploymentService) initialiseK8sClient(conf *config.Config, logger log.Logger) (*k8s.Manager, error) {
	clientset, err := k8s.NewClientset(conf.Kubernetes.ClusterURL, conf.Kubernetes.Namespace)
	if err != nil {
		return nil, err
	}

	kubeOperator, err := k8s.NewOperator(k8s.WithLogger(logger))
	if err != nil {
		return nil, err
	}

	return k8s.NewManager(conf, logger, clientset, kubeOperator)
}

func (d *DeploymentService) Deploy() error {
	if err := d.initializeClient(d.conf, d.logger); err != nil {
		return err
	}

	return d.client.Deploy()
}

func (d *DeploymentService) Destroy() error {
	if err := d.initializeClient(d.conf, d.logger); err != nil {
		return err
	}

	return d.client.Destroy()
}
