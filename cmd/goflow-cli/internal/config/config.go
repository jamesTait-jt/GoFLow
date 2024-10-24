package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GoFlowServer GoFlowServer `yaml:"goflow_server"`
	Workerpool   Workerpool   `yaml:"workerpool"`
	Redis        Redis        `yaml:"redis"`
	Kubernetes   Kubernetes   `yaml:"kubernetes"`
}

type GoFlowServer struct {
	Image    string `yaml:"image"`
	Replicas int32  `yaml:"replicas"`
	Port     int32  `yaml:"port"`
}

type Workerpool struct {
	Image              string `yaml:"image"`
	Replicas           int    `yaml:"replicas"`
	PathToHandlers     string `yaml:"path_to_handlers"`
	PluginBuilderImage string `yaml:"plugin_builder_image"`
}

type Redis struct {
	Image    string `yaml:"image"`
	Replicas int32  `yaml:"replicas"`
}

type Kubernetes struct {
	Namespace  string `yaml:"namespace"`
	ClusterURL string `yaml:"clusterUrl"`
}

func LoadConfig(filePath string) (*Config, error) {
	yamlFile, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("error reading .goflow.yaml: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		return nil, fmt.Errorf("error parsing .goflow.yaml: %v", err)
	}

	return &config, nil
}

var GoFlowHostPort = "30000"

var DockerNetworkID = "goflow-network"

var GoflowImage = "goflow-server:latest"
var GoflowContainerName = "goflow-server"

var RedisImage = "redis:latest"
var RedisContainerName = "goflow-redis-server"

var PluginBuilderImage = "plugin-builder"

var WorkerpoolImage = "workerpool"
var WorkerpoolContainerName = "goflow-workerpool"
