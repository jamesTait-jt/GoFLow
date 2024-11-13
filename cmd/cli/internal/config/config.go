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

// Config represents the structure of the configuration file for the Goflow application.
// It consolidates settings for the GoFlow server, worker pool, and message broker. As
// well as cluster level Kubernetes information so that the CLI can connect to it.
type Config struct {
	GoFlowServer GoFlowServer `yaml:"goflow_server"`
	Workerpool   Workerpool   `yaml:"workerpool"`
	Redis        Redis        `yaml:"redis"`
	Kubernetes   Kubernetes   `yaml:"kubernetes"`
}

// GoFlowServer defines the configuration settings for the GoFlow server component.
type GoFlowServer struct {
	// Image is the Docker image for the goflow server container
	// (usually goflow-server:latest).
	Image string `yaml:"image"`

	// Replicas is the number of server instances that you want. These will be placed
	// behind a load balancer.
	Replicas int32 `yaml:"replicas"`

	// Address is the address of the load balancer through which the server can be
	// reached from outside of the Kubernetes cluster. This will be provisioned in
	// the deployment.
	Address string `yaml:"address"`
}

// Workerpool defines the configuration for the worker pool component of the Goflow
// application.
type Workerpool struct {
	// Image is the Docker image for the workerpool container
	// (usually goflow-workerpool:latest).
	Image string `yaml:"image"`

	// PluginBuilderImage is the Docker image for the container that is used to build
	// the plugins for the workerpool so that it can handle custom tasks. This will be
	// run as an InitContainer so that the workerpools will have the compiled plugins
	// ready to load in.
	PluginBuilderImage string `yaml:"plugin_builder_image"`

	// Replicas is the number of workerpool instances you would like. All of these will
	// be configured to concurrently listen to the message broker.
	Replicas int32 `yaml:"replicas"`

	// PathToHandlers is the path to the directory containing the uncompiled go plugins.
	// This currently must be a location on the Kubernetes master node.
	//
	// If using Minikube, you can load them onto the node via `minikube scp` and then
	//  point to the path you copied them to using this config option.
	PathToHandlers string `yaml:"path_to_handlers"`
}

// Redis defines the configuration for the Redis message brokers. This only needs to be
// populated if redis is the chosen message broker.
type Redis struct {
	// Image is the Docker image for redis
	Image string `yaml:"image"`

	// Replicas is the number of replicas you want to have to the redis instance
	Replicas int32 `yaml:"replicas"`
}

// Kubernetes defines cluuster level configuration.
type Kubernetes struct {
	// Namespace is the namespace that you would like to deploy the Kubernetes resources
	// under (example `goflow`).
	Namespace string `yaml:"namespace"`

	// ClusterURL is the url of the Kubernetes cluster that you would like to deploy GoFlow
	// to
	ClusterURL string `yaml:"clusterUrl"`
}

// Load reads and parses the configuration file at the specified (relative or absolute) file path.
// It unmarshals the YAML content into the Config struct and assigns it to a package-level variable.
// Returns an error if the file cannot be read or if there is a parsing issue.
func Load(filePath string) error {
	yamlFile, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("error reading .goflow.yaml: %v", err)
	}

	var innerConfig Config

	err = yaml.Unmarshal(yamlFile, &innerConfig)

	config = &innerConfig

	if err != nil {
		return fmt.Errorf("error parsing .goflow.yaml: %v", err)
	}

	return nil
}

// Get retrieves the loaded configuration as a Config struct pointer.
// Returns an error if the configuration has not been loaded successfully.
func Get() (*Config, error) {
	if config == nil {
		return nil, errors.New("config has not been loaded")
	}

	return config, nil
}
