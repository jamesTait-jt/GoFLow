package k8s

import (
	"errors"
	"testing"
	"time"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/resource"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/watch"
)

func Test_NewDeploymentManager(t *testing.T) {
	t.Run("Initialises a new deployment manager", func(t *testing.T) {
		// Arrange
		conf := new(config.Config)
		logger := new(log.TestifyMock)
		clientset := new(mockClientset)

		expectedDeploymentManager := &DeploymentManager{
			logger:          logger,
			resourceBuilder: resource.NewBuilder(conf, clientset),
			executor:        NewDeploymentExecutor(logger),
		}

		// Act
		d := NewDeploymentManager(conf, logger, clientset)

		// Assert
		assert.Equal(t, d, expectedDeploymentManager)
	})
}

func Test_Deployer_DeployNamespace(t *testing.T) {
	t.Run("Builds the namespace resource and executes the apply command", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		returnedResource := &resource.Resource{}
		builder.On("Build", resource.Namespace).Once().Return(returnedResource)
		executor.On("ApplyAndWait", returnedResource, timeout).Once().Return(nil)

		// Act
		err := d.DeployNamespace()

		// Assert
		assert.NoError(t, err)
		builder.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if execute command failed", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		namespaceResource := &resource.Resource{}
		builder.On("Build", resource.Namespace).Once().Return(namespaceResource)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", namespaceResource, timeout).Once().Return(execErr)

		// Act
		err := d.DeployNamespace()

		// Assert
		assert.EqualError(t, err, execErr.Error())
	})
}

func Test_Deployer_DeployMessageBroker(t *testing.T) {
	t.Run("Builds the message broker resources and executes the apply command for each of them", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		deploymentResource := &resource.Resource{}
		builder.On("Build", resource.MessageBrokerDeployment).Once().Return(deploymentResource)
		executor.On("ApplyAndWait", deploymentResource, timeout).Once().Return(nil)

		serviceResource := &resource.Resource{}
		builder.On("Build", resource.MessageBrokerService).Once().Return(serviceResource)
		executor.On("ApplyAndWait", serviceResource, timeout).Once().Return(nil)

		// Act
		err := d.DeployMessageBroker()

		// Assert
		assert.NoError(t, err)
		builder.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if the deployment resource failed to apply", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		deploymentResource := &resource.Resource{}
		builder.On("Build", resource.MessageBrokerDeployment).Once().Return(deploymentResource)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", deploymentResource, timeout).Once().Return(execErr)

		// Act
		err := d.DeployMessageBroker()

		// Assert
		assert.EqualError(t, err, execErr.Error())
	})

	t.Run("Returns an error if the deployment resource failed to apply", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		deploymentResource := &resource.Resource{}
		builder.On("Build", resource.MessageBrokerDeployment).Once().Return(deploymentResource)
		executor.On("ApplyAndWait", deploymentResource, timeout).Once().Return(nil)

		serviceResource := &resource.Resource{}
		builder.On("Build", resource.MessageBrokerService).Once().Return(serviceResource)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", serviceResource, timeout).Once().Return(execErr)

		// Act
		err := d.DeployMessageBroker()

		// Assert
		assert.EqualError(t, err, execErr.Error())
	})
}

func Test_Deployer_DeployGRPCServer(t *testing.T) {
	t.Run("Builds the gRPC server resources and executes the apply command for each of them", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		deploymentResource := &resource.Resource{}
		builder.On("Build", resource.GRPCServerDeployment).Once().Return(deploymentResource)
		executor.On("ApplyAndWait", deploymentResource, timeout).Once().Return(nil)

		serviceResource := &resource.Resource{}
		builder.On("Build", resource.GRPCServerService).Once().Return(serviceResource)
		executor.On("ApplyAndWait", serviceResource, timeout).Once().Return(nil)

		// Act
		err := d.DeployGRPCServer()

		// Assert
		assert.NoError(t, err)
		builder.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if the deployment resource failed to apply", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		deploymentService := &resource.Resource{}
		builder.On("Build", resource.GRPCServerDeployment).Once().Return(deploymentService)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", deploymentService, timeout).Once().Return(execErr)

		// Act
		err := d.DeployGRPCServer()

		// Assert
		assert.EqualError(t, err, execErr.Error())
	})

	t.Run("Returns an error if the service resource failed to apply", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		deploymentResource := &resource.Resource{}
		builder.On("Build", resource.GRPCServerDeployment).Once().Return(deploymentResource)
		executor.On("ApplyAndWait", deploymentResource, timeout).Once().Return(nil)

		serviceResource := &resource.Resource{}
		builder.On("Build", resource.GRPCServerService).Once().Return(serviceResource)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", serviceResource, timeout).Once().Return(execErr)

		// Act
		err := d.DeployGRPCServer()

		// Assert
		assert.EqualError(t, err, execErr.Error())
	})
}

func Test_Deployer_DeployWorkerpools(t *testing.T) {
	t.Run("Builds the workerpool resources and executes the apply command for each of them", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		pvResource := &resource.Resource{}
		builder.On("Build", resource.WorkerpoolPV).Once().Return(pvResource)
		executor.On("ApplyAndWait", pvResource, timeout).Once().Return(nil)

		pvcResource := &resource.Resource{}
		builder.On("Build", resource.WorkerpoolPVC).Once().Return(pvcResource)
		executor.On("ApplyAndWait", pvcResource, timeout).Once().Return(nil)

		deploymentResource := &resource.Resource{}
		builder.On("Build", resource.WorkerpoolDeployment).Once().Return(deploymentResource)
		executor.On("ApplyAndWait", deploymentResource, timeout).Once().Return(nil)

		// Act
		err := d.DeployWorkerpools()

		// Assert
		assert.NoError(t, err)
		builder.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if the PV resource failed to apply", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		pvResource := &resource.Resource{}
		builder.On("Build", resource.WorkerpoolPV).Once().Return(pvResource)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", pvResource, timeout).Once().Return(execErr)

		// Act
		err := d.DeployWorkerpools()

		// Assert
		assert.EqualError(t, err, execErr.Error())
	})

	t.Run("Returns an error if the PVC resource failed to apply", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		pvResource := &resource.Resource{}
		builder.On("Build", resource.WorkerpoolPV).Once().Return(pvResource)
		executor.On("ApplyAndWait", pvResource, timeout).Once().Return(nil)

		pvcResource := &resource.Resource{}
		builder.On("Build", resource.WorkerpoolPVC).Once().Return(pvcResource)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", pvcResource, timeout).Once().Return(execErr)

		// Act
		err := d.DeployWorkerpools()

		// Assert
		assert.EqualError(t, err, execErr.Error())
	})

	t.Run("Returns an error if the workerpool deployment resource failed to apply", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor}

		pvResource := &resource.Resource{}
		builder.On("Build", resource.WorkerpoolPV).Once().Return(pvResource)
		executor.On("ApplyAndWait", pvResource, timeout).Once().Return(nil)

		pvcResource := &resource.Resource{}
		builder.On("Build", resource.WorkerpoolPVC).Once().Return(pvcResource)
		executor.On("ApplyAndWait", pvcResource, timeout).Once().Return(nil)

		deploymentResource := &resource.Resource{}
		builder.On("Build", resource.WorkerpoolDeployment).Once().Return(deploymentResource)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", deploymentResource, timeout).Once().Return(execErr)

		// Act
		err := d.DeployWorkerpools()

		// Assert
		assert.EqualError(t, err, execErr.Error())
	})
}

func Test_Deployer_DestroyAll(t *testing.T) {
	t.Run("Deletes the namespace and persistent volume and returns no error", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)
		logger := new(log.TestifyMock)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor, logger: logger}

		namespaceResource := &resource.Resource{}
		builder.On("Build", resource.Namespace).Once().Return(namespaceResource)
		executor.On("DestroyAndWait", namespaceResource, timeout).Once().Return(nil)

		pvResource := &resource.Resource{}
		builder.On("Build", resource.WorkerpoolPV).Once().Return(pvResource)
		executor.On("DestroyAndWait", pvResource, timeout).Once().Return(nil)

		logger.On("Success", "GoFlow destroyed!").Once()

		// Act
		err := d.DestroyAll()

		// Assert
		assert.NoError(t, err)
		logger.AssertExpectations(t)
		builder.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if deleting the namespace fails", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)
		logger := new(log.TestifyMock)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor, logger: logger}

		namespaceResource := &resource.Resource{}
		builder.On("Build", resource.Namespace).Once().Return(namespaceResource)

		execErr := errors.New("destroy namespace error")
		executor.On("DestroyAndWait", namespaceResource, timeout).Once().Return(execErr)

		// Act
		err := d.DestroyAll()

		// Assert
		assert.EqualError(t, err, execErr.Error())
		logger.AssertExpectations(t)
		builder.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if deleting the persistent volume (PV) fails", func(t *testing.T) {
		// Arrange
		builder := new(mockResourceBuilder)
		executor := new(mockDeploymentExecutor)
		logger := new(log.TestifyMock)

		d := &DeploymentManager{resourceBuilder: builder, executor: executor, logger: logger}

		namespaceResource := &resource.Resource{}
		builder.On("Build", resource.Namespace).Once().Return(namespaceResource)
		executor.On("DestroyAndWait", namespaceResource, timeout).Once().Return(nil)

		pvResource := &resource.Resource{}
		builder.On("Build", resource.WorkerpoolPV).Once().Return(pvResource)

		execErr := errors.New("destroy persistent volume error")
		executor.On("DestroyAndWait", pvResource, timeout).Once().Return(execErr)

		// Act
		err := d.DestroyAll()

		// Assert
		assert.EqualError(t, err, execErr.Error())
		logger.AssertExpectations(t)
		builder.AssertExpectations(t)
		executor.AssertExpectations(t)
	})
}

type mockOperator struct {
	mock.Mock
}

func (m *mockOperator) Apply(kubeResource ApplyGetter) (bool, error) {
	args := m.Called(kubeResource)
	return args.Bool(0), args.Error(1)
}

func (m *mockOperator) Delete(kubeResource Deleter) (bool, error) {
	args := m.Called(kubeResource)
	return args.Bool(0), args.Error(1)
}

func (m *mockOperator) WaitFor(kubeResource Watcher, eventTypes []watch.EventType, timeout time.Duration) error {
	args := m.Called(kubeResource, eventTypes, timeout)
	return args.Error(0)
}

type mockClientset struct {
	mock.Mock
}

func (m *mockClientset) Namespaces() resource.NamespaceInterface {
	args := m.Called()
	return args.Get(0).(resource.NamespaceInterface)
}

func (m *mockClientset) Deployments() resource.DeploymentInterface {
	args := m.Called()
	return args.Get(0).(resource.DeploymentInterface)
}

func (m *mockClientset) Services() resource.ServiceInterface {
	args := m.Called()
	return args.Get(0).(resource.ServiceInterface)
}

func (m *mockClientset) PersistentVolumes() resource.PersistentVolumeInterface {
	args := m.Called()
	return args.Get(0).(resource.PersistentVolumeInterface)
}

func (m *mockClientset) PersistentVolumeClaims() resource.PersistentVolumeClaimInterface {
	args := m.Called()
	return args.Get(0).(resource.PersistentVolumeClaimInterface)
}

type mockResourceBuilder struct {
	mock.Mock
}

func (m *mockResourceBuilder) Build(resourceKey resource.Key) *resource.Resource {
	args := m.Called(resourceKey)
	return args.Get(0).(*resource.Resource)
}

type mockDeploymentExecutor struct {
	mock.Mock
}

func (m *mockDeploymentExecutor) ApplyAndWait(kubeResource identifiableWatchableApplyGetter, timeout time.Duration) error {
	args := m.Called(kubeResource, timeout)
	return args.Error(0)
}

func (m *mockDeploymentExecutor) DestroyAndWait(kubeResource identifiableWatchableDeleter, timeout time.Duration) error {
	args := m.Called(kubeResource, timeout)
	return args.Error(0)
}
