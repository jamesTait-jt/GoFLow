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

func Deploy(handlersPath string, local bool) error {
	if local {
		return deployLocal(handlersPath)
	}

	return deployKubernetes(handlersPath)
}

func deployLocal(handlersPath string) error {
	dockerClient, err := docker.New()
	if err != nil {
		return fmt.Errorf("error creating Docker client: %v", err)
	}
	defer dockerClient.Close()

	fmt.Println("Creating Docker network...")

	if err := dockerClient.CreateNetwork(config.DockerNetworkID); err != nil {
		return err
	}

	fmt.Println("Starting goflow gRPC service...")

	if err := startGoflowService(dockerClient); err != nil {
		return err
	}

	fmt.Println("Starting Redis container...")

	if err := startRedis(dockerClient); err != nil {
		return err
	}

	fmt.Println("Compiling plugins...")

	if err := compilePlugins(dockerClient, handlersPath); err != nil {
		return err
	}

	fmt.Println("Starting WorkerPool container...")

	if err := startWorkerPool(dockerClient, handlersPath); err != nil {
		return err
	}

	fmt.Println("Deployment successful!")

	return nil
}

func deployKubernetes(handlersPath string) error {
	conf, err := config.LoadConfig("/Users/James.Tait/go/src/github.com/jamesTait-jt/goflow/.goflow.yaml")
	if err != nil {
		return err
	}

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

	if err = kubeClient.CreateOrReplacePV(workerpool.HandlersPV(conf)); err != nil {
		return err
	}

	if err = kubeClient.CreateOrReplacePVC(workerpool.HandlersPVC(conf)); err != nil {
		return err
	}

	if err = kubeClient.CreateOrReplaceAndRunJob(workerpool.Job(conf)); err != nil {
		return err
	}

	return nil
}

func startGoflowService(dockerClient *docker.Docker) error {
	containerID, err := dockerClient.CreateContainer(
		&container.Config{
			Image: config.GoflowImage,
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				"50051/tcp": []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: config.GoFlowHostPort,
					},
				},
			},
		},
		config.DockerNetworkID,
		config.GoflowContainerName,
	)

	if err != nil {
		return fmt.Errorf("failed to create goflow container: %v", err)
	}

	if err := dockerClient.StartContainer(containerID); err != nil {
		return fmt.Errorf("error starting goflow container: %v", err)
	}

	fmt.Println("goflow container started successfully")

	return nil
}

func startRedis(dockerClient *docker.Docker) error {
	containerInfo, err := dockerClient.ContainerInfo(config.RedisContainerName)
	if err != nil {
		return err
	}

	if containerInfo.Running {
		fmt.Println("Redis container already started")

		return nil
	}

	if !containerInfo.Exists {
		err = dockerClient.PullImage(config.RedisImage)
		if err != nil {
			return fmt.Errorf("failed to pull redis image: %v", err)
		}

		containerInfo.ID, err = dockerClient.CreateContainer(
			&container.Config{
				Image: config.RedisImage,
			},
			nil,
			config.DockerNetworkID,
			config.RedisContainerName,
		)

		if err != nil {
			return fmt.Errorf("error creating Redis container: %v", err)
		}
	}

	if err = dockerClient.StartContainer(containerInfo.ID); err != nil {
		return fmt.Errorf("error starting redis container: %v", err)
	}

	fmt.Println("Redis container started successfully")

	return nil
}

func compilePlugins(dockerClient *docker.Docker, handlersPath string) error {
	containerID, err := dockerClient.CreateContainer(
		&container.Config{
			Image: config.PluginBuilderImage,
			Cmd:   []string{"handlers"},
		},
		&container.HostConfig{
			Binds:      []string{fmt.Sprintf("%s:/app/handlers", handlersPath)},
			AutoRemove: true,
		},
		"",
		"",
	)

	if err != nil {
		return fmt.Errorf("failed to create plugin-builder container: %v", err)
	}

	if err = dockerClient.StartContainer(containerID); err != nil {
		return fmt.Errorf("failed to start plugin-builder container: %v", err)
	}

	err = dockerClient.WaitForContainerToFinish(containerID)
	if err != nil {
		return fmt.Errorf("failed to wait for plugin-builder to finish: %v", err)
	}

	containerPassed, err := dockerClient.ContainerPassed(containerID)
	if err != nil {
		return fmt.Errorf("failed to check plugin-builder exit status: %v", err)
	}

	if !containerPassed {
		return errors.New("plugin-builder container failed")
	}

	fmt.Println("plugins compiled!")

	return nil
}

func startWorkerPool(dockerClient *docker.Docker, handlersPath string) error {
	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/app/handlers", handlersPath)},
	}

	containerID, err := dockerClient.CreateContainer(
		&container.Config{
			Image: config.WorkerpoolImage,
			Cmd: []string{
				"--broker-type", "redis",
				"--broker-addr", fmt.Sprintf("%s:6379", config.RedisContainerName),
				"--handlers-path", "/app/handlers/compiled",
			},
		},
		hostConfig,
		config.DockerNetworkID,
		config.WorkerpoolContainerName,
	)

	if err != nil {
		return fmt.Errorf("failed to create workerpool container: %v", err)
	}

	if err := dockerClient.StartContainer(containerID); err != nil {
		return fmt.Errorf("error starting workerpool container: %v", err)
	}

	fmt.Println("WorkerPool container started successfully")

	return nil
}
