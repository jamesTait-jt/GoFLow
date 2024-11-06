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
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_NewDeploymentManager(t *testing.T) {
	t.Run("Initialises a new deployment manager", func(t *testing.T) {
		// Arrange
		conf := new(config.Config)
		logger := new(log.TestifyMock)

		clientset := &Clients{
			clientset: fake.NewSimpleClientset(),
		}

		expectedDeploymentManager := &DeploymentManager{
			logger:          logger,
			configMapper:    NewConfigMapper(conf),
			resourceFactory: resource.NewFactory(clientset),
			executor:        NewDeploymentExecutor(logger),
		}

		// Act
		d := NewDeploymentManager(conf, logger, clientset)

		// Assert
		assert.Equal(t, d, expectedDeploymentManager)
	})
}

func Test_DeploymentManager_DeployNamespace(t *testing.T) {
	t.Run("Builds the namespace resource and executes the apply command", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		namespaceConf := acapiv1.Namespace("namespace")
		configMapper.On("GetNamespaceConfig", resource.Namespace).Once().Return(namespaceConf, nil)

		namespace := &resource.Resource{}
		resourceFactory.On("CreateNamespace", namespaceConf).Once().Return(namespace)

		executor.On("ApplyAndWait", namespace, timeout).Once().Return(nil)

		// Act
		err := d.DeployNamespace()

		// Assert
		assert.NoError(t, err)
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't get apply config", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		getConfErr := errors.New("couldnt get conf")
		configMapper.On("GetNamespaceConfig", resource.Namespace).Once().Return(nil, getConfErr)

		// Act
		err := d.DeployNamespace()

		// Assert
		assert.EqualError(t, err, getConfErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if execute command failed", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		namespaceConf := acapiv1.Namespace("namespace")
		configMapper.On("GetNamespaceConfig", resource.Namespace).Once().Return(namespaceConf, nil)

		namespace := &resource.Resource{}
		resourceFactory.On("CreateNamespace", namespaceConf).Once().Return(namespace)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", namespace, timeout).Once().Return(execErr)

		// Act
		err := d.DeployNamespace()

		// Assert
		assert.EqualError(t, err, execErr.Error())
	})
}

func Test_DeploymentManager_DeployMessageBroker(t *testing.T) {
	t.Run("Builds the deployment and service resources and executes the apply commands", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		deploymentConf := acappsv1.Deployment("message-broker-deployment", "test-namespace")
		configMapper.On("GetDeploymentConfig", resource.MessageBrokerDeployment).Once().Return(deploymentConf, nil)

		deployment := &resource.Resource{}
		resourceFactory.On("CreateDeployment", deploymentConf).Once().Return(deployment)

		executor.On("ApplyAndWait", deployment, timeout).Once().Return(nil)

		serviceConf := acapiv1.Service("message-broker-service", "test-namespace")
		configMapper.On("GetServiceConfig", resource.MessageBrokerService).Once().Return(serviceConf, nil)

		service := &resource.Resource{}
		resourceFactory.On("CreateService", serviceConf).Once().Return(service)

		executor.On("ApplyAndWait", service, timeout).Once().Return(nil)

		// Act
		err := d.DeployMessageBroker()

		// Assert
		assert.NoError(t, err)
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't get deployment config", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		getConfErr := errors.New("couldn't get deployment config")
		configMapper.On("GetDeploymentConfig", resource.MessageBrokerDeployment).Once().Return(nil, getConfErr)

		// Act
		err := d.DeployMessageBroker()

		// Assert
		assert.EqualError(t, err, getConfErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't get service config", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		deploymentConf := acappsv1.Deployment("message-broker-deployment", "test-namespace")
		configMapper.On("GetDeploymentConfig", resource.MessageBrokerDeployment).Once().Return(deploymentConf, nil)

		deployment := &resource.Resource{}
		resourceFactory.On("CreateDeployment", deploymentConf).Once().Return(deployment)

		executor.On("ApplyAndWait", deployment, timeout).Once().Return(nil)

		getConfErr := errors.New("couldn't get service config")
		configMapper.On("GetServiceConfig", resource.MessageBrokerService).Once().Return(nil, getConfErr)

		// Act
		err := d.DeployMessageBroker()

		// Assert
		assert.EqualError(t, err, getConfErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if execute command for deployment failed", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		deploymentConf := acappsv1.Deployment("message-broker-deployment", "test-namespace")
		configMapper.On("GetDeploymentConfig", resource.MessageBrokerDeployment).Once().Return(deploymentConf, nil)

		deployment := &resource.Resource{}
		resourceFactory.On("CreateDeployment", deploymentConf).Once().Return(deployment)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", deployment, timeout).Once().Return(execErr)

		// Act
		err := d.DeployMessageBroker()

		// Assert
		assert.EqualError(t, err, execErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if execute command for service failed", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		deploymentConf := acappsv1.Deployment("message-broker-deployment", "test-namespace")
		configMapper.On("GetDeploymentConfig", resource.MessageBrokerDeployment).Once().Return(deploymentConf, nil)

		deployment := &resource.Resource{}
		resourceFactory.On("CreateDeployment", deploymentConf).Once().Return(deployment)

		executor.On("ApplyAndWait", deployment, timeout).Once().Return(nil)

		serviceConf := acapiv1.Service("message-broker-service", "test-namespace")
		configMapper.On("GetServiceConfig", resource.MessageBrokerService).Once().Return(serviceConf, nil)

		service := &resource.Resource{}
		resourceFactory.On("CreateService", serviceConf).Once().Return(service)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", service, timeout).Once().Return(execErr)

		// Act
		err := d.DeployMessageBroker()

		// Assert
		assert.EqualError(t, err, execErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})
}
func Test_DeploymentManager_DeployGRPCServer(t *testing.T) {
	t.Run("Builds the deployment and service resources and executes the apply commands", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		deploymentConf := acappsv1.Deployment("grpc-server-deployment", "test-namespace")
		configMapper.On("GetDeploymentConfig", resource.GRPCServerDeployment).Once().Return(deploymentConf, nil)

		deployment := &resource.Resource{}
		resourceFactory.On("CreateDeployment", deploymentConf).Once().Return(deployment)

		executor.On("ApplyAndWait", deployment, timeout).Once().Return(nil)

		serviceConf := acapiv1.Service("grpc-server-service", "test-namespace")
		configMapper.On("GetServiceConfig", resource.GRPCServerService).Once().Return(serviceConf, nil)

		service := &resource.Resource{}
		resourceFactory.On("CreateService", serviceConf).Once().Return(service)

		executor.On("ApplyAndWait", service, timeout).Once().Return(nil)

		// Act
		err := d.DeployGRPCServer()

		// Assert
		assert.NoError(t, err)
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't get deployment config", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		getConfErr := errors.New("couldn't get deployment config")
		configMapper.On("GetDeploymentConfig", resource.GRPCServerDeployment).Once().Return(nil, getConfErr)

		// Act
		err := d.DeployGRPCServer()

		// Assert
		assert.EqualError(t, err, getConfErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't get service config", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		deploymentConf := acappsv1.Deployment("grpc-server-deployment", "test-namespace")
		configMapper.On("GetDeploymentConfig", resource.GRPCServerDeployment).Once().Return(deploymentConf, nil)

		deployment := &resource.Resource{}
		resourceFactory.On("CreateDeployment", deploymentConf).Once().Return(deployment)

		executor.On("ApplyAndWait", deployment, timeout).Once().Return(nil)

		getConfErr := errors.New("couldn't get service config")
		configMapper.On("GetServiceConfig", resource.GRPCServerService).Once().Return(nil, getConfErr)

		// Act
		err := d.DeployGRPCServer()

		// Assert
		assert.EqualError(t, err, getConfErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if execute command for deployment failed", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		deploymentConf := acappsv1.Deployment("grpc-server-deployment", "test-namespace")
		configMapper.On("GetDeploymentConfig", resource.GRPCServerDeployment).Once().Return(deploymentConf, nil)

		deployment := &resource.Resource{}
		resourceFactory.On("CreateDeployment", deploymentConf).Once().Return(deployment)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", deployment, timeout).Once().Return(execErr)

		// Act
		err := d.DeployGRPCServer()

		// Assert
		assert.EqualError(t, err, execErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if execute command for service failed", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		deploymentConf := acappsv1.Deployment("grpc-server-deployment", "test-namespace")
		configMapper.On("GetDeploymentConfig", resource.GRPCServerDeployment).Once().Return(deploymentConf, nil)

		deployment := &resource.Resource{}
		resourceFactory.On("CreateDeployment", deploymentConf).Once().Return(deployment)

		executor.On("ApplyAndWait", deployment, timeout).Once().Return(nil)

		serviceConf := acapiv1.Service("grpc-server-service", "test-namespace")
		configMapper.On("GetServiceConfig", resource.GRPCServerService).Once().Return(serviceConf, nil)

		service := &resource.Resource{}
		resourceFactory.On("CreateService", serviceConf).Once().Return(service)

		execErr := errors.New("apply and wait err")
		executor.On("ApplyAndWait", service, timeout).Once().Return(execErr)

		// Act
		err := d.DeployGRPCServer()

		// Assert
		assert.EqualError(t, err, execErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})
}

func Test_DeploymentManager_DeployWorkerpools(t *testing.T) {
	t.Run("Builds the PV, PVC, and deployment resources and executes the apply commands", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		pvConf := acapiv1.PersistentVolume("workerpool-pv")
		configMapper.On("GetPersistentVolumeConfig", resource.WorkerpoolPV).Once().Return(pvConf, nil)

		pv := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolume", pvConf).Once().Return(pv)

		executor.On("ApplyAndWait", pv, timeout).Once().Return(nil)

		pvcConf := acapiv1.PersistentVolumeClaim("workerpool-pvc", "test-namespace")
		configMapper.On("GetPersistentVolumeClaimConfig", resource.WorkerpoolPVC).Once().Return(pvcConf, nil)

		pvc := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolumeClaim", pvcConf).Once().Return(pvc)

		executor.On("ApplyAndWait", pvc, timeout).Once().Return(nil)

		deploymentConf := acappsv1.Deployment("workerpool-deployment", "test-namespace")
		configMapper.On("GetDeploymentConfig", resource.WorkerpoolDeployment).Once().Return(deploymentConf, nil)

		deployment := &resource.Resource{}
		resourceFactory.On("CreateDeployment", deploymentConf).Once().Return(deployment)

		executor.On("ApplyAndWait", deployment, timeout).Once().Return(nil)

		// Act
		err := d.DeployWorkerpools()

		// Assert
		assert.NoError(t, err)
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't get persistent volume config", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		getConfErr := errors.New("couldn't get persistent volume config")
		configMapper.On("GetPersistentVolumeConfig", resource.WorkerpoolPV).Once().Return(nil, getConfErr)

		// Act
		err := d.DeployWorkerpools()

		// Assert
		assert.EqualError(t, err, getConfErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't get persistent volume claim config", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		pvConf := acapiv1.PersistentVolume("workerpool-pv")
		configMapper.On("GetPersistentVolumeConfig", resource.WorkerpoolPV).Once().Return(pvConf, nil)

		pv := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolume", pvConf).Once().Return(pv)

		executor.On("ApplyAndWait", pv, timeout).Once().Return(nil)

		getConfErr := errors.New("couldn't get persistent volume claim config")
		configMapper.On("GetPersistentVolumeClaimConfig", resource.WorkerpoolPVC).Once().Return(nil, getConfErr)

		// Act
		err := d.DeployWorkerpools()

		// Assert
		assert.EqualError(t, err, getConfErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't get deployment config", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		pvConf := acapiv1.PersistentVolume("workerpool-pv")
		configMapper.On("GetPersistentVolumeConfig", resource.WorkerpoolPV).Once().Return(pvConf, nil)

		pv := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolume", pvConf).Once().Return(pv)

		executor.On("ApplyAndWait", pv, timeout).Once().Return(nil)

		pvcConf := acapiv1.PersistentVolumeClaim("workerpool-pvc", "test-namespace")
		configMapper.On("GetPersistentVolumeClaimConfig", resource.WorkerpoolPVC).Once().Return(pvcConf, nil)

		pvc := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolumeClaim", pvcConf).Once().Return(pvc)

		executor.On("ApplyAndWait", pvc, timeout).Once().Return(nil)

		getConfErr := errors.New("couldn't get deployment config")
		configMapper.On("GetDeploymentConfig", resource.WorkerpoolDeployment).Once().Return(nil, getConfErr)

		// Act
		err := d.DeployWorkerpools()

		// Assert
		assert.EqualError(t, err, getConfErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if execute command for persistent volume failed", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		pvConf := acapiv1.PersistentVolume("workerpool-pv")
		configMapper.On("GetPersistentVolumeConfig", resource.WorkerpoolPV).Once().Return(pvConf, nil)

		pv := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolume", pvConf).Once().Return(pv)

		execErr := errors.New("apply and wait err for pv")
		executor.On("ApplyAndWait", pv, timeout).Once().Return(execErr)

		// Act
		err := d.DeployWorkerpools()

		// Assert
		assert.EqualError(t, err, execErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if execute command for persistent volume claim failed", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		pvConf := acapiv1.PersistentVolume("workerpool-pv")
		configMapper.On("GetPersistentVolumeConfig", resource.WorkerpoolPV).Once().Return(pvConf, nil)

		pv := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolume", pvConf).Once().Return(pv)

		executor.On("ApplyAndWait", pv, timeout).Once().Return(nil)

		pvcConf := acapiv1.PersistentVolumeClaim("workerpool-pvc", "test-namespace")
		configMapper.On("GetPersistentVolumeClaimConfig", resource.WorkerpoolPVC).Once().Return(pvcConf, nil)

		pvc := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolumeClaim", pvcConf).Once().Return(pvc)

		execErr := errors.New("apply and wait err for pvc")
		executor.On("ApplyAndWait", pvc, timeout).Once().Return(execErr)

		// Act
		err := d.DeployWorkerpools()

		// Assert
		assert.EqualError(t, err, execErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("Returns an error if execute command for deployment failed", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor}

		pvConf := acapiv1.PersistentVolume("workerpool-pv")
		configMapper.On("GetPersistentVolumeConfig", resource.WorkerpoolPV).Once().Return(pvConf, nil)

		pv := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolume", pvConf).Once().Return(pv)

		executor.On("ApplyAndWait", pv, timeout).Once().Return(nil)

		pvcConf := acapiv1.PersistentVolumeClaim("workerpool-pvc", "test-namespace")
		configMapper.On("GetPersistentVolumeClaimConfig", resource.WorkerpoolPVC).Once().Return(pvcConf, nil)

		pvc := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolumeClaim", pvcConf).Once().Return(pvc)

		executor.On("ApplyAndWait", pvc, timeout).Once().Return(nil)

		deploymentConf := acappsv1.Deployment("workerpool-deployment", "test-namespace")
		configMapper.On("GetDeploymentConfig", resource.WorkerpoolDeployment).Once().Return(deploymentConf, nil)

		deployment := &resource.Resource{}
		resourceFactory.On("CreateDeployment", deploymentConf).Once().Return(deployment)

		execErr := errors.New("apply and wait err for deployment")
		executor.On("ApplyAndWait", deployment, timeout).Once().Return(execErr)

		// Act
		err := d.DeployWorkerpools()

		// Assert
		assert.EqualError(t, err, execErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
	})
}

func Test_DeploymentManager_DestroyAll(t *testing.T) {
	t.Run("Destroys the namespace and persistent volume successfully", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)
		logger := new(log.TestifyMock)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor, logger: logger}

		namespaceConf := acapiv1.Namespace("test-namespace")
		configMapper.On("GetNamespaceConfig", resource.Namespace).Once().Return(namespaceConf, nil)

		namespace := &resource.Resource{}
		resourceFactory.On("CreateNamespace", namespaceConf).Once().Return(namespace)

		executor.On("DestroyAndWait", namespace, timeout).Once().Return(nil)

		pvConf := acapiv1.PersistentVolume("workerpool-pv")
		configMapper.On("GetPersistentVolumeConfig", resource.WorkerpoolPV).Once().Return(pvConf, nil)

		pv := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolume", pvConf).Once().Return(pv)

		executor.On("DestroyAndWait", pv, timeout).Once().Return(nil)

		logger.On("Success", "GoFlow destroyed!").Once()

		// Act
		err := d.DestroyAll()

		// Assert
		assert.NoError(t, err)
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't get namespace config", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)
		logger := new(log.TestifyMock)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor, logger: logger}

		getConfErr := errors.New("couldn't get namespace config")
		configMapper.On("GetNamespaceConfig", resource.Namespace).Once().Return(nil, getConfErr)

		// Act
		err := d.DestroyAll()

		// Assert
		assert.EqualError(t, err, getConfErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't destroy the namespace", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)
		logger := new(log.TestifyMock)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor, logger: logger}

		namespaceConf := acapiv1.Namespace("test-namespace")
		configMapper.On("GetNamespaceConfig", resource.Namespace).Once().Return(namespaceConf, nil)

		namespace := &resource.Resource{}
		resourceFactory.On("CreateNamespace", namespaceConf).Once().Return(namespace)

		execErr := errors.New("destroy and wait err for namespace")
		executor.On("DestroyAndWait", namespace, timeout).Once().Return(execErr)

		// Act
		err := d.DestroyAll()

		// Assert
		assert.EqualError(t, err, execErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't get persistent volume config", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)
		logger := new(log.TestifyMock)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor, logger: logger}

		namespaceConf := acapiv1.Namespace("test-namespace")
		configMapper.On("GetNamespaceConfig", resource.Namespace).Once().Return(namespaceConf, nil)

		namespace := &resource.Resource{}
		resourceFactory.On("CreateNamespace", namespaceConf).Once().Return(namespace)

		executor.On("DestroyAndWait", namespace, timeout).Once().Return(nil)

		getConfErr := errors.New("couldn't get persistent volume config")
		configMapper.On("GetPersistentVolumeConfig", resource.WorkerpoolPV).Once().Return(nil, getConfErr)

		// Act
		err := d.DestroyAll()

		// Assert
		assert.EqualError(t, err, getConfErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("Returns an error if couldn't destroy the persistent volume", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)
		logger := new(log.TestifyMock)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor, logger: logger}

		namespaceConf := acapiv1.Namespace("test-namespace")
		configMapper.On("GetNamespaceConfig", resource.Namespace).Once().Return(namespaceConf, nil)

		namespace := &resource.Resource{}
		resourceFactory.On("CreateNamespace", namespaceConf).Once().Return(namespace)

		executor.On("DestroyAndWait", namespace, timeout).Once().Return(nil)

		pvConf := acapiv1.PersistentVolume("workerpool-pv")
		configMapper.On("GetPersistentVolumeConfig", resource.WorkerpoolPV).Once().Return(pvConf, nil)

		pv := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolume", pvConf).Once().Return(pv)

		execErr := errors.New("destroy and wait err for pv")
		executor.On("DestroyAndWait", pv, timeout).Once().Return(execErr)

		// Act
		err := d.DestroyAll()

		// Assert
		assert.EqualError(t, err, execErr.Error())
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("Logs success when all resources are destroyed", func(t *testing.T) {
		// Arrange
		configMapper := new(mockConfigMapper)
		resourceFactory := new(mockResourceFactory)
		executor := new(mockDeploymentExecutor)
		logger := new(log.TestifyMock)

		d := &DeploymentManager{configMapper: configMapper, resourceFactory: resourceFactory, executor: executor, logger: logger}

		// Namespace configuration
		namespaceConf := acapiv1.Namespace("test-namespace")
		configMapper.On("GetNamespaceConfig", resource.Namespace).Once().Return(namespaceConf, nil)

		namespace := &resource.Resource{}
		resourceFactory.On("CreateNamespace", namespaceConf).Once().Return(namespace)

		executor.On("DestroyAndWait", namespace, timeout).Once().Return(nil)

		// Persistent Volume configuration
		pvConf := acapiv1.PersistentVolume("workerpool-pv")
		configMapper.On("GetPersistentVolumeConfig", resource.WorkerpoolPV).Once().Return(pvConf, nil)

		pv := &resource.Resource{}
		resourceFactory.On("CreatePersistentVolume", pvConf).Once().Return(pv)

		executor.On("DestroyAndWait", pv, timeout).Once().Return(nil)

		logger.On("Success", "GoFlow destroyed!").Once()

		// Act
		err := d.DestroyAll()

		// Assert
		assert.NoError(t, err)
		configMapper.AssertExpectations(t)
		resourceFactory.AssertExpectations(t)
		executor.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}

type mockConfigMapper struct {
	mock.Mock
}

func (m *mockConfigMapper) GetNamespaceConfig(resourceKey resource.Key) (*acapiv1.NamespaceApplyConfiguration, error) {
	args := m.Called(resourceKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*acapiv1.NamespaceApplyConfiguration), args.Error(1)
}

func (m *mockConfigMapper) GetDeploymentConfig(resourceKey resource.Key) (*acappsv1.DeploymentApplyConfiguration, error) {
	args := m.Called(resourceKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*acappsv1.DeploymentApplyConfiguration), args.Error(1)
}

func (m *mockConfigMapper) GetServiceConfig(resourceKey resource.Key) (*acapiv1.ServiceApplyConfiguration, error) {
	args := m.Called(resourceKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*acapiv1.ServiceApplyConfiguration), args.Error(1)
}

func (m *mockConfigMapper) GetPersistentVolumeConfig(resourceKey resource.Key) (*acapiv1.PersistentVolumeApplyConfiguration, error) {
	args := m.Called(resourceKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*acapiv1.PersistentVolumeApplyConfiguration), args.Error(1)
}

func (m *mockConfigMapper) GetPersistentVolumeClaimConfig(resourceKey resource.Key) (*acapiv1.PersistentVolumeClaimApplyConfiguration, error) {
	args := m.Called(resourceKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*acapiv1.PersistentVolumeClaimApplyConfiguration), args.Error(1)
}

type mockResourceFactory struct {
	mock.Mock
}

func (m *mockResourceFactory) CreateNamespace(config *acapiv1.NamespaceApplyConfiguration) *resource.Resource {
	args := m.Called(config)
	return args.Get(0).(*resource.Resource)
}

func (m *mockResourceFactory) CreateDeployment(config *acappsv1.DeploymentApplyConfiguration) *resource.Resource {
	args := m.Called(config)
	return args.Get(0).(*resource.Resource)
}

func (m *mockResourceFactory) CreateService(config *acapiv1.ServiceApplyConfiguration) *resource.Resource {
	args := m.Called(config)
	return args.Get(0).(*resource.Resource)
}

func (m *mockResourceFactory) CreatePersistentVolume(config *acapiv1.PersistentVolumeApplyConfiguration) *resource.Resource {
	args := m.Called(config)
	return args.Get(0).(*resource.Resource)
}

func (m *mockResourceFactory) CreatePersistentVolumeClaim(config *acapiv1.PersistentVolumeClaimApplyConfiguration) *resource.Resource {
	args := m.Called(config)
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
