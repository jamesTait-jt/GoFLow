package k8s

import (
	"time"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/resource"
	"github.com/jamesTait-jt/goflow/pkg/log"
)

var timeout = 30 * time.Second

type resourceBuilder interface {
	Build(resourceKey resource.Key) *resource.Resource
}

type deploymentExecutor interface {
	ApplyAndWait(kubeResource identifiableWatchableApplyGetter, timeout time.Duration) error
	DestroyAndWait(kubeResource identifiableWatchableDeleter, timeout time.Duration) error
}

type DeploymentManager struct {
	logger          log.Logger
	resourceBuilder resourceBuilder
	executor        deploymentExecutor
}

func NewDeploymentManager(conf *config.Config, logger log.Logger, clients resource.Clientset) *DeploymentManager {
	return &DeploymentManager{
		logger:          logger,
		resourceBuilder: resource.NewBuilder(conf, clients),
		executor:        NewDeploymentExecutor(logger),
	}
}

func (d *DeploymentManager) DeployNamespace() error {
	namespace := d.resourceBuilder.Build(resource.Namespace)

	return d.executor.ApplyAndWait(namespace, timeout)
}

func (d *DeploymentManager) DeployMessageBroker() error {
	deployment := d.resourceBuilder.Build(resource.MessageBrokerDeployment)
	if err := d.executor.ApplyAndWait(deployment, timeout); err != nil {
		return err
	}

	service := d.resourceBuilder.Build(resource.MessageBrokerService)

	return d.executor.ApplyAndWait(service, timeout)
}

func (d *DeploymentManager) DeployGRPCServer() error {
	deployment := d.resourceBuilder.Build(resource.GRPCServerDeployment)
	if err := d.executor.ApplyAndWait(deployment, timeout); err != nil {
		return err
	}

	service := d.resourceBuilder.Build(resource.GRPCServerService)

	return d.executor.ApplyAndWait(service, timeout)
}

func (d *DeploymentManager) DeployWorkerpools() error {
	pv := d.resourceBuilder.Build(resource.WorkerpoolPV)
	if err := d.executor.ApplyAndWait(pv, timeout); err != nil {
		return err
	}

	pvc := d.resourceBuilder.Build(resource.WorkerpoolPVC)
	if err := d.executor.ApplyAndWait(pvc, timeout); err != nil {
		return err
	}

	deployment := d.resourceBuilder.Build(resource.WorkerpoolDeployment)

	return d.executor.ApplyAndWait(deployment, timeout)
}

func (d *DeploymentManager) DestroyAll() error {
	// Deleting the namespace will delete all associated resources
	namespace := d.resourceBuilder.Build(resource.Namespace)
	if err := d.executor.DestroyAndWait(namespace, timeout); err != nil {
		return err
	}

	// Persistent volumes are not scoped to a namespace
	pv := d.resourceBuilder.Build(resource.WorkerpoolPV)
	if err := d.executor.DestroyAndWait(pv, timeout); err != nil {
		return err
	}

	d.logger.Success("GoFlow destroyed!")

	return nil
}
