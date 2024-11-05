package service

import "fmt"

type deployer interface {
	DeployNamespace() error
	DeployMessageBroker() error
	DeployGRPCServer() error
	DeployWorkerpools() error
	DestroyAll() error
}

type DeploymentService struct {
	deployer deployer
}

func NewDeploymentService(d deployer) *DeploymentService {
	return &DeploymentService{
		deployer: d,
	}
}

func (d *DeploymentService) Deploy() error {
	if err := d.deployer.DeployNamespace(); err != nil {
		return fmt.Errorf("failed to deploy namespace: %w", err)
	}

	if err := d.deployer.DeployMessageBroker(); err != nil {
		return fmt.Errorf("failed to deploy message broker: %w", err)
	}

	if err := d.deployer.DeployGRPCServer(); err != nil {
		return fmt.Errorf("failed to deploy gRPC server: %w", err)
	}

	if err := d.deployer.DeployWorkerpools(); err != nil {
		return fmt.Errorf("failed to deploy worker pools: %w", err)
	}

	return nil
}

func (d *DeploymentService) Destroy() error {
	return d.deployer.DestroyAll()
}
