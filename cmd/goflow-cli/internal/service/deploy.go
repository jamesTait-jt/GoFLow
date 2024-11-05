package service

import "fmt"

type deploymentManager interface {
	DeployNamespace() error
	DeployMessageBroker() error
	DeployGRPCServer() error
	DeployWorkerpools() error
	DestroyAll() error
}

type DeploymentService struct {
	deploymentManager deploymentManager
}

func NewDeploymentService(d deploymentManager) *DeploymentService {
	return &DeploymentService{
		deploymentManager: d,
	}
}

func (d *DeploymentService) Deploy() error {
	if err := d.deploymentManager.DeployNamespace(); err != nil {
		return fmt.Errorf("failed to deploy namespace: %w", err)
	}

	if err := d.deploymentManager.DeployMessageBroker(); err != nil {
		return fmt.Errorf("failed to deploy message broker: %w", err)
	}

	if err := d.deploymentManager.DeployGRPCServer(); err != nil {
		return fmt.Errorf("failed to deploy gRPC server: %w", err)
	}

	if err := d.deploymentManager.DeployWorkerpools(); err != nil {
		return fmt.Errorf("failed to deploy worker pools: %w", err)
	}

	return nil
}

func (d *DeploymentService) Destroy() error {
	return d.deploymentManager.DestroyAll()
}
