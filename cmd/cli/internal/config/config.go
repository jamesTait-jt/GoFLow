package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// config will be set when the LoadConfig function is called
var config *Config

type Config struct {
	GoFlowServer GoFlowServer `yaml:"goflow_server"`
	Workerpool   Workerpool   `yaml:"workerpool"`
	Redis        Redis        `yaml:"redis"`
	Kubernetes   *Kubernetes  `yaml:"kubernetes"`
	Docker       *Docker      `yaml:"docker"`
}

type GoFlowServer struct {
	Image    string `yaml:"image"`
	Replicas int32  `yaml:"replicas"`
	IP       string `yaml:"address"`
	Port     int32  `yaml:"port"`
}

type Workerpool struct {
	Image              string `yaml:"image"`
	Replicas           int32  `yaml:"replicas"`
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

type Docker struct {
	NetworkID string `yaml:"network_id"`
}

func Load(filePath string) error {
	yamlFile, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("error reading .goflow.yaml: %v", err)
	}

	var innerConfig Config

	err = yaml.Unmarshal(yamlFile, &innerConfig)
	if err != nil {
		return fmt.Errorf("error parsing .goflow.yaml: %v", err)
	}

	if err = ValidateDeploymentEnvironment(innerConfig); err != nil {
		return err
	}

	config = &innerConfig

	return nil
}

func Get() (*Config, error) {
	if config == nil {
		return nil, errors.New("config has not been loaded")
	}

	return config, nil
}
