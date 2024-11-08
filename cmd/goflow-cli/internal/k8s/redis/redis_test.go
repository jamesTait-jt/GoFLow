//go:build unit

package redis

import (
	"testing"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestDeployment(t *testing.T) {
	t.Run("Initialises the deployment correctly", func(t *testing.T) {
		// Arrange
		conf := &config.Config{
			Kubernetes: config.Kubernetes{
				Namespace: "test-namespace",
			},
			Redis: config.Redis{
				Replicas: int32(3),
				Image:    "test-image",
			},
		}

		// Act
		deployment := Deployment(conf)

		// Assert
		assert.Equal(t, deploymentName, *deployment.Name)
		assert.Equal(t, conf.Kubernetes.Namespace, *deployment.Namespace)
		assert.Equal(t, labels, deployment.Labels)

		assert.NotNil(t, deployment.Spec)
		assert.Equal(t, conf.Redis.Replicas, *deployment.Spec.Replicas)

		assert.NotNil(t, deployment.Spec.Selector)
		assert.Equal(t, labels, deployment.Spec.Selector.MatchLabels)

		assert.NotNil(t, deployment.Spec.Template)
		assert.Equal(t, labels, deployment.Spec.Template.Labels)

		assert.NotNil(t, deployment.Spec.Template.Spec)
		assert.Len(t, deployment.Spec.Template.Spec.Containers, 1)

		container := deployment.Spec.Template.Spec.Containers[0]
		assert.Equal(t, deploymentContainerName, *container.Name)
		assert.Equal(t, conf.Redis.Image, *container.Image)

		assert.Len(t, container.Ports, 1)
		port := container.Ports[0]
		assert.Equal(t, apiv1.ProtocolTCP, *port.Protocol)
		assert.Equal(t, RedisPort, *port.ContainerPort)
	})
}

func TestService(t *testing.T) {
	t.Run("Initialises the service correctly", func(t *testing.T) {
		// Arrange
		conf := &config.Config{
			Kubernetes: config.Kubernetes{
				Namespace: "test-namespace",
			},
		}

		// Act
		service := Service(conf)

		// Assert
		assert.Equal(t, ServiceName, *service.Name)
		assert.Equal(t, conf.Kubernetes.Namespace, *service.Namespace)
		assert.Equal(t, labels, service.Labels)

		assert.NotNil(t, service.Spec)
		assert.Equal(t, apiv1.ServiceTypeClusterIP, *service.Spec.Type)

		assert.NotNil(t, service.Spec.Selector)
		assert.Equal(t, labels, service.Spec.Selector)

		assert.NotNil(t, service.Spec.Ports)
		assert.Len(t, service.Spec.Ports, 1)

		port := service.Spec.Ports[0]
		assert.Equal(t, RedisPort, *port.Port)
		assert.Equal(t, intstr.FromInt32(RedisPort), *port.TargetPort)
	})
}
