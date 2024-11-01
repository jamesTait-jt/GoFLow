package grpcserver

import (
	"testing"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Test_Deployment(t *testing.T) {
	t.Run("Initialises the deployment correctly", func(t *testing.T) {
		// Arrange
		conf := &config.Config{
			Kubernetes: config.Kubernetes{
				Namespace: "test-namespace",
			},
			GoFlowServer: config.GoFlowServer{
				Replicas: 3,
				Image:    "test-image",
			},
		}

		// Act
		deploymentConfig := Deployment(conf)

		// Assert
		assert.NotNil(t, deploymentConfig)
		assert.Equal(t, deploymentName, *deploymentConfig.Name)
		assert.Equal(t, conf.Kubernetes.Namespace, *deploymentConfig.Namespace)
		assert.Equal(t, labels, deploymentConfig.Labels)

		assert.NotNil(t, deploymentConfig.Spec)
		assert.Equal(t, conf.GoFlowServer.Replicas, *deploymentConfig.Spec.Replicas)

		assert.NotNil(t, deploymentConfig.Spec.Selector)
		assert.Equal(t, labels, deploymentConfig.Spec.Selector.MatchLabels)

		assert.NotNil(t, deploymentConfig.Spec.Template)
		assert.Equal(t, labels, deploymentConfig.Spec.Template.Labels)

		assert.NotNil(t, deploymentConfig.Spec.Template.Spec)
		assert.Len(t, deploymentConfig.Spec.Template.Spec.Containers, 1)
		container := deploymentConfig.Spec.Template.Spec.Containers[0]
		assert.Equal(t, deploymentContainerName, *container.Name)
		assert.Equal(t, conf.GoFlowServer.Image, *container.Image)
		assert.Equal(t, apiv1.PullNever, *container.ImagePullPolicy)

		assert.Len(t, container.Ports, 1)
		port := container.Ports[0]
		assert.Equal(t, apiv1.ProtocolTCP, *port.Protocol)
		assert.Equal(t, grpcPort, *port.ContainerPort)
	})
}

func Test_Service(t *testing.T) {
	t.Run("Initialises the service correctly", func(t *testing.T) {
		// Arrange
		conf := &config.Config{
			Kubernetes: config.Kubernetes{
				Namespace: "test-namespace",
			},
			GoFlowServer: config.GoFlowServer{
				Address: "1.2.3.4",
			},
		}

		// Act
		serviceConfig := Service(conf)

		// Assert
		assert.NotNil(t, serviceConfig)
		assert.Equal(t, serviceName, *serviceConfig.Name)
		assert.Equal(t, conf.Kubernetes.Namespace, *serviceConfig.Namespace)
		assert.Equal(t, labels, serviceConfig.Labels)

		assert.NotNil(t, serviceConfig.Spec)
		assert.Equal(t, labels, serviceConfig.Spec.Selector)
		assert.Equal(t, apiv1.ServiceTypeLoadBalancer, *serviceConfig.Spec.Type)
		assert.Equal(t, conf.GoFlowServer.Address, *serviceConfig.Spec.LoadBalancerIP)

		assert.Len(t, serviceConfig.Spec.Ports, 1)
		port := serviceConfig.Spec.Ports[0]
		assert.Equal(t, grpcPort, *port.Port)
		assert.Equal(t, intstr.FromInt32(grpcPort), *port.TargetPort)
	})
}
