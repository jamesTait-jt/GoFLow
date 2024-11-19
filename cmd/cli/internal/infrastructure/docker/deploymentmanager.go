package docker

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/infrastructure"
	"github.com/jamesTait-jt/goflow/pkg/log"
)

type DeploymentManager struct {
	conf   *config.Config
	logger log.Logger
	client *Docker
}

func NewDeploymentManager(conf *config.Config, logger log.Logger) (*DeploymentManager, error) {
	dockerClient, err := New()
	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %v", err)
	}

	return &DeploymentManager{
		conf:   conf,
		logger: logger,
		client: dockerClient,
	}, nil
}

func (d *DeploymentManager) DeployNamespace() error {
	return d.client.CreateNetwork(d.conf.Docker.NetworkID)
}

func (d *DeploymentManager) DeployMessageBroker() error {
	containerInfo, err := d.client.ContainerInfo(infrastructure.RedisContainerName)
	if err != nil {
		return err
	}

	if containerInfo.Running {
		fmt.Println("Redis container already started")

		return nil
	}

	if !containerInfo.Exists {
		err = d.client.PullImage(d.conf.Redis.Image)
		if err != nil {
			return fmt.Errorf("failed to pull redis image: %v", err)
		}

		containerInfo.ID, err = d.client.CreateContainer(
			&container.Config{
				Image: d.conf.Redis.Image,
			},
			nil,
			d.conf.Docker.NetworkID,
			infrastructure.RedisContainerName,
		)

		if err != nil {
			return fmt.Errorf("error creating Redis container: %v", err)
		}
	}

	if err = d.client.StartContainer(containerInfo.ID); err != nil {
		return fmt.Errorf("error starting redis container: %v", err)
	}

	fmt.Println("Redis container started successfully")

	return nil
}

func (d *DeploymentManager) DeployGRPCServer() error {
	return nil
}

func (d *DeploymentManager) DeployWorkerpools() error {
	return nil
}

func (d *DeploymentManager) DestroyAll() error {
	return nil
}
