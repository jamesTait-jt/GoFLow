//go:build unit

package k8s

import (
	"testing"

	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/infrastructure/k8s/grpcserver"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/infrastructure/k8s/redis"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/infrastructure/k8s/resource"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/infrastructure/k8s/workerpool"
	"github.com/stretchr/testify/assert"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

func Test_NewConfigMapper(t *testing.T) {
	t.Run("Initialises config mapper", func(t *testing.T) {
		// Arrange
		conf := new(config.Config)

		expectedConfigMapper := &ConfigMapper{conf: conf}

		// Act
		c := NewConfigMapper(conf)

		// Assert
		assert.Equal(t, expectedConfigMapper, c)
	})
}

func Test_ConfigMapper_GetNamespaceConfig(t *testing.T) {
	t.Run("Returns correct namespace config", func(t *testing.T) {
		// Arrange
		conf := &config.Config{
			Kubernetes: &config.Kubernetes{
				Namespace: "namespace",
			},
		}
		cm := &ConfigMapper{conf: conf}

		expectedNamespace := acapiv1.Namespace(cm.conf.Kubernetes.Namespace)

		// Act
		applyConf, err := cm.GetNamespaceConfig(resource.Namespace)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedNamespace, applyConf)
	})

	t.Run("Returns error if unrecognised key", func(t *testing.T) {
		// Arrange
		conf := new(config.Config)
		cm := &ConfigMapper{conf: conf}

		// Act
		applyConf, err := cm.GetNamespaceConfig(-1)

		// Assert
		assert.Nil(t, applyConf)
		assert.EqualError(t, err, ("didnt recognise namespace resource key '-1'"))
	})
}

func Test_ConfigMapper_GetDeploymentConfig(t *testing.T) {
	t.Run("Returns correct deployment config for key", func(t *testing.T) {
		type test struct {
			inKey           resource.Key
			wantApplyConfig *acappsv1.DeploymentApplyConfiguration
		}

		conf := &config.Config{
			Kubernetes: &config.Kubernetes{
				Namespace: "namespace",
			},
		}

		tts := []test{
			{resource.GRPCServerDeployment, grpcserver.Deployment(conf)},
			{resource.MessageBrokerDeployment, redis.Deployment(conf)},
			{resource.WorkerpoolDeployment, workerpool.Deployment(conf)},
		}

		for _, tt := range tts {
			t.Run(tt.inKey.String(), func(t *testing.T) {
				// Arrange
				cm := &ConfigMapper{conf: conf}

				// Act
				applyConf, err := cm.GetDeploymentConfig(tt.inKey)

				// Assert
				assert.NoError(t, err)
				assert.Equal(t, tt.wantApplyConfig, applyConf)
			})
		}
	})

	t.Run("Returns error if unrecognised key", func(t *testing.T) {
		// Arrange
		conf := new(config.Config)
		cm := &ConfigMapper{conf: conf}

		// Act
		applyConf, err := cm.GetDeploymentConfig(-1)

		// Assert
		assert.Nil(t, applyConf)
		assert.EqualError(t, err, ("didnt recognise deployment resource key '-1'"))
	})
}

func Test_ConfigMapper_GetServiceConfig(t *testing.T) {
	t.Run("Returns correct service config for key", func(t *testing.T) {
		type test struct {
			inKey           resource.Key
			wantApplyConfig *acapiv1.ServiceApplyConfiguration
		}

		conf := &config.Config{
			Kubernetes: &config.Kubernetes{
				Namespace: "namespace",
			},
		}

		tts := []test{
			{resource.GRPCServerService, grpcserver.Service(conf)},
			{resource.MessageBrokerService, redis.Service(conf)},
		}

		for _, tt := range tts {
			t.Run(tt.inKey.String(), func(t *testing.T) {
				// Arrange
				cm := &ConfigMapper{conf: conf}

				// Act
				applyConf, err := cm.GetServiceConfig(tt.inKey)

				// Assert
				assert.NoError(t, err)
				assert.Equal(t, tt.wantApplyConfig, applyConf)
			})
		}
	})

	t.Run("Returns error if unrecognised key", func(t *testing.T) {
		// Arrange
		conf := new(config.Config)
		cm := &ConfigMapper{conf: conf}

		// Act
		applyConf, err := cm.GetServiceConfig(-1)

		// Assert
		assert.Nil(t, applyConf)
		assert.EqualError(t, err, ("didnt recognise service resource key '-1'"))
	})
}

func Test_ConfigMapper_GetPersistentVolumeConfig(t *testing.T) {
	t.Run("Returns correct pv config", func(t *testing.T) {
		conf := new(config.Config)
		// Arrange
		cm := &ConfigMapper{conf: conf}

		expectedPV := workerpool.HandlersPV(conf)

		// Act
		applyConf, err := cm.GetPersistentVolumeConfig(resource.WorkerpoolPV)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedPV, applyConf)
	})

	t.Run("Returns error if unrecognised key", func(t *testing.T) {
		// Arrange
		conf := new(config.Config)
		cm := &ConfigMapper{conf: conf}

		// Act
		applyConf, err := cm.GetPersistentVolumeConfig(-1)

		// Assert
		assert.Nil(t, applyConf)
		assert.EqualError(t, err, ("didnt recognise persistent volume resource key '-1'"))
	})
}

func Test_ConfigMapper_GetPersistentVolumeClaimConfig(t *testing.T) {
	t.Run("Returns correct pvc config", func(t *testing.T) {
		// Arrange
		conf := &config.Config{
			Kubernetes: &config.Kubernetes{
				Namespace: "namespace",
			},
		}
		cm := &ConfigMapper{conf: conf}

		expectedPVC := workerpool.HandlersPVC(conf)

		// Act
		applyConf, err := cm.GetPersistentVolumeClaimConfig(resource.WorkerpoolPVC)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedPVC, applyConf)
	})

	t.Run("Returns error if unrecognised key", func(t *testing.T) {
		// Arrange
		conf := &config.Config{
			Kubernetes: &config.Kubernetes{
				Namespace: "namespace",
			},
		}
		cm := &ConfigMapper{conf: conf}

		// Act
		applyConf, err := cm.GetPersistentVolumeClaimConfig(-1)

		// Assert
		assert.Nil(t, applyConf)
		assert.EqualError(t, err, ("didnt recognise persistent volume claim resource key '-1'"))
	})
}
