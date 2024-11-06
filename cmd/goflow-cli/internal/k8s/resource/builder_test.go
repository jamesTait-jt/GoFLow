package resource

import (
	"testing"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/grpcserver"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/redis"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/workerpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

func Test_NewBuilder(t *testing.T) {
	t.Run("Initialises the builder", func(t *testing.T) {
		// Arrange
		conf := new(config.Config)
		clientset := new(mockClientset)

		expectedBuilder := &Builder{
			clients:              clientset,
			namespaceApplyConfig: acapiv1.Namespace(conf.Kubernetes.Namespace),
			messageBrokerApplyConfigs: messageBrokerApplyConfigs{
				deployment: redis.Deployment(conf),
				service:    redis.Service(conf),
			},
			gRPCServerApplyConfigs: gRPCServerApplyConfigs{
				deployment: grpcserver.Deployment(conf),
				service:    grpcserver.Service(conf),
			},
			workerpoolApplyConfigs: workerpoolApplyConfigs{
				pv:         workerpool.HandlersPV(conf),
				pvc:        workerpool.HandlersPVC(conf),
				deployment: workerpool.Deployment(conf),
			},
		}

		// Act
		b := NewBuilder(conf, clientset)

		// Assert
		assert.Equal(t, expectedBuilder, b)
	})
}

type mockClientset struct {
	mock.Mock
}

func (m *mockClientset) Namespaces() NamespaceInterface {
	args := m.Called()
	return args.Get(0).(NamespaceInterface)
}

func (m *mockClientset) Deployments() DeploymentInterface {
	args := m.Called()
	return args.Get(0).(DeploymentInterface)
}

func (m *mockClientset) Services() ServiceInterface {
	args := m.Called()
	return args.Get(0).(ServiceInterface)
}

func (m *mockClientset) PersistentVolumes() PersistentVolumeInterface {
	args := m.Called()
	return args.Get(0).(PersistentVolumeInterface)
}

func (m *mockClientset) PersistentVolumeClaims() PersistentVolumeClaimInterface {
	args := m.Called()
	return args.Get(0).(PersistentVolumeClaimInterface)
}
