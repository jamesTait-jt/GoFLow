//go:build unit

package workerpool

import (
	"fmt"
	"testing"

	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/k8s/redis"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestHandlersPV(t *testing.T) {
	t.Run("Initialises the persistent volume correctly", func(t *testing.T) {
		// Arrange
		conf := &config.Config{
			Workerpool: config.Workerpool{
				PathToHandlers: "/path/to/handlers",
			},
			Kubernetes: config.Kubernetes{
				Namespace: "test-namespace",
			},
		}

		// Act
		pv := HandlersPV(conf)

		// Assert
		assert.Equal(t, pvName, *pv.Name)

		assert.NotNil(t, pv.Spec)
		assert.Len(t, pv.Spec.AccessModes, 1)
		assert.Equal(t, apiv1.ReadWriteMany, pv.Spec.AccessModes[0])
		assert.Equal(t, storageClassName, *pv.Spec.StorageClassName)

		assert.NotNil(t, pv.Spec.Capacity)
		assert.Len(t, *pv.Spec.Capacity, 1)
		assert.Equal(t, resource.MustParse("1Gi"), (*pv.Spec.Capacity)[apiv1.ResourceStorage])

		assert.NotNil(t, pv.Spec.HostPath)
		assert.Equal(t, conf.Workerpool.PathToHandlers, *pv.Spec.HostPath.Path)
	})
}

func TestHandlersPVC(t *testing.T) {
	t.Run("Initialises the persistent volume claim correctly", func(t *testing.T) {
		// Arrange
		conf := &config.Config{
			Kubernetes: config.Kubernetes{
				Namespace: "test-namespace",
			},
		}

		// Act
		pvc := HandlersPVC(conf)

		// Assert
		assert.Equal(t, pvcName, *pvc.Name)
		assert.Equal(t, conf.Kubernetes.Namespace, *pvc.Namespace)

		assert.NotNil(t, pvc.Spec)
		assert.Equal(t, pvName, *pvc.Spec.VolumeName)
		assert.Len(t, pvc.Spec.AccessModes, 1)
		assert.Equal(t, apiv1.ReadWriteMany, pvc.Spec.AccessModes[0])
		assert.Equal(t, storageClassName, *pvc.Spec.StorageClassName)

		assert.NotNil(t, pvc.Spec.Resources)
		assert.NotNil(t, pvc.Spec.Resources.Requests)
		assert.Len(t, *pvc.Spec.Resources.Requests, 1)
		assert.Equal(t, resource.MustParse("1Gi"), (*pvc.Spec.Resources.Requests)[apiv1.ResourceStorage])
	})
}

func TestDeployment(t *testing.T) {
	t.Run("Initialises the deployment correctly", func(t *testing.T) {
		// Arrange
		conf := &config.Config{
			Kubernetes: config.Kubernetes{
				Namespace: "test-namespace",
			},
			Workerpool: config.Workerpool{
				Replicas:           int32(3),
				Image:              "workerpool-image",
				PluginBuilderImage: "plugin-builder-image",
			},
		}

		// Act
		deployment := Deployment(conf)

		// Assert
		assert.Equal(t, deploymentName, *deployment.Name)
		assert.Equal(t, conf.Kubernetes.Namespace, *deployment.Namespace)
		assert.Equal(t, labels, deployment.Labels)

		assert.NotNil(t, deployment.Spec)
		assert.Equal(t, conf.Workerpool.Replicas, *deployment.Spec.Replicas)

		assert.NotNil(t, deployment.Spec.Selector)
		assert.Equal(t, labels, deployment.Spec.Selector.MatchLabels)

		assert.NotNil(t, deployment.Spec.Template)
		assert.Equal(t, labels, deployment.Spec.Template.Labels)

		assert.NotNil(t, deployment.Spec.Template.Spec)
		assert.Equal(t, apiv1.RestartPolicyAlways, *deployment.Spec.Template.Spec.RestartPolicy)

		assert.Len(t, deployment.Spec.Template.Spec.Volumes, 1)
		volume := deployment.Spec.Template.Spec.Volumes[0]
		assert.Equal(t, volumeMountName, *volume.Name)
		assert.NotNil(t, volume.PersistentVolumeClaim)
		assert.Equal(t, pvcName, *volume.PersistentVolumeClaim.ClaimName)

		assert.Len(t, deployment.Spec.Template.Spec.InitContainers, 1)
		initContainer := deployment.Spec.Template.Spec.InitContainers[0]
		assert.Equal(t, pluginBuilderContainerName, *initContainer.Name)
		assert.Equal(t, conf.Workerpool.PluginBuilderImage, *initContainer.Image)
		assert.Equal(t, apiv1.PullNever, *initContainer.ImagePullPolicy)
		assert.Equal(t, []string{"/app/handlers"}, initContainer.Args)

		assert.Len(t, initContainer.VolumeMounts, 1)
		assert.Equal(t, volumeMountName, *initContainer.VolumeMounts[0].Name)
		assert.Equal(t, "/app/handlers", *initContainer.VolumeMounts[0].MountPath)

		assert.Len(t, deployment.Spec.Template.Spec.Containers, 1)
		workerpoolContainer := deployment.Spec.Template.Spec.Containers[0]
		assert.Equal(t, workerpoolContainerName, *workerpoolContainer.Name)
		assert.Equal(t, conf.Workerpool.Image, *workerpoolContainer.Image)
		assert.Equal(t, apiv1.PullNever, *workerpoolContainer.ImagePullPolicy)
		assert.Equal(t, []string{
			"--broker-type", "redis",
			"--broker-addr", fmt.Sprintf("%s:%d", redis.ServiceName, redis.RedisPort),
			"--handlers-path", "/app/handlers/compiled",
		}, workerpoolContainer.Args)

		assert.Len(t, workerpoolContainer.VolumeMounts, 1)
		assert.Equal(t, volumeMountName, *workerpoolContainer.VolumeMounts[0].Name)
		assert.Equal(t, "/app/handlers", *workerpoolContainer.VolumeMounts[0].MountPath)
	})
}
