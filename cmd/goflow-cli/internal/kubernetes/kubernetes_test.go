package kubernetes

import (
	"context"
	"errors"
	"testing"

	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Test_New(t *testing.T) {
	t.Run("Initialises KubeClient", func(t *testing.T) {
		// Arrange
		mockConfigBuilder := new(MockConfigBuilder)
		mockKubeClientBuilder := new(MockKubeClientBuilder)
		mockLogger := new(log.TestifyMock)

		clusterURL := "cluster.url"
		kubeConfigPath := "path/to/kube/config"
		kubeConfig := &rest.Config{}
		clientSet := &kubernetes.Clientset{}

		mockConfigBuilder.On("GetKubeConfigPath").Return(kubeConfigPath, nil)
		mockConfigBuilder.On("BuildConfig", clusterURL, kubeConfigPath).Return(kubeConfig, nil)
		mockKubeClientBuilder.On("NewForConfig", kubeConfig).Return(clientSet, nil)

		// Act
		kube, err := New(
			clusterURL,
			WithConfigBuilder(mockConfigBuilder),
			WithKubeClientBuilder(mockKubeClientBuilder),
			WithLogger(mockLogger),
		)

		// Assert
		mockConfigBuilder.AssertExpectations(t)
		mockKubeClientBuilder.AssertExpectations(t)

		assert.Nil(t, err)
		assert.NotNil(t, kube.ctx)
		assert.NotNil(t, kube.waiter)
		assert.Equal(t, clientSet, kube.client)
		assert.Equal(t, mockLogger, kube.logger)
	})

	t.Run("Returns error if could not find kube config", func(t *testing.T) {
		// Arrange
		mockConfigBuilder := new(MockConfigBuilder)
		mockKubeClientBuilder := new(MockKubeClientBuilder)
		mockLogger := new(log.TestifyMock)

		clusterURL := "cluster.url"

		configBuilderErr := errors.New("conf builder err")
		mockConfigBuilder.On("GetKubeConfigPath").Return("", configBuilderErr)

		// Act
		kube, err := New(
			clusterURL,
			WithConfigBuilder(mockConfigBuilder),
			WithKubeClientBuilder(mockKubeClientBuilder),
			WithLogger(mockLogger),
		)

		// Assert
		mockConfigBuilder.AssertExpectations(t)
		mockConfigBuilder.AssertNotCalled(t, "BuildConfig")
		mockKubeClientBuilder.AssertNotCalled(t, "NewForConfig")

		assert.EqualError(t, err, configBuilderErr.Error())
		assert.Nil(t, kube)
	})

	t.Run("Returns error if could not build kube config", func(t *testing.T) {
		// Arrange
		mockConfigBuilder := new(MockConfigBuilder)
		mockKubeClientBuilder := new(MockKubeClientBuilder)
		mockLogger := new(log.TestifyMock)

		clusterURL := "cluster.url"
		kubeConfigPath := "path/to/kube/config"
		configBuilderErr := errors.New("conf builder err")

		mockConfigBuilder.On("GetKubeConfigPath").Return(kubeConfigPath, nil)
		mockConfigBuilder.On("BuildConfig", clusterURL, kubeConfigPath).Return(nil, configBuilderErr)

		// Act
		kube, err := New(
			clusterURL,
			WithConfigBuilder(mockConfigBuilder),
			WithKubeClientBuilder(mockKubeClientBuilder),
			WithLogger(mockLogger),
		)

		// Assert
		mockConfigBuilder.AssertExpectations(t)
		mockKubeClientBuilder.AssertNotCalled(t, "NewForConfig")

		assert.EqualError(t, err, configBuilderErr.Error())
		assert.Nil(t, kube)
	})

	t.Run("Returns error if could not create Clientset", func(t *testing.T) {
		// Arrange
		mockConfigBuilder := new(MockConfigBuilder)
		mockKubeClientBuilder := new(MockKubeClientBuilder)
		mockLogger := new(log.TestifyMock)

		clusterURL := "cluster.url"
		kubeConfigPath := "path/to/kube/config"
		kubeConfig := &rest.Config{}

		kubeClientBuilderErr := errors.New("kube client builder err")

		mockConfigBuilder.On("GetKubeConfigPath").Return(kubeConfigPath, nil)
		mockConfigBuilder.On("BuildConfig", clusterURL, kubeConfigPath).Return(kubeConfig, nil)
		mockKubeClientBuilder.On("NewForConfig", kubeConfig).Return(nil, kubeClientBuilderErr)

		// Act
		kube, err := New(
			clusterURL,
			WithConfigBuilder(mockConfigBuilder),
			WithKubeClientBuilder(mockKubeClientBuilder),
			WithLogger(mockLogger),
		)

		// Assert
		mockConfigBuilder.AssertExpectations(t)
		mockKubeClientBuilder.AssertExpectations(t)

		assert.EqualError(t, err, kubeClientBuilderErr.Error())
		assert.Nil(t, kube)
	})
}

func Test_KubeClient_ApplyDeployment(t *testing.T) {
	t.Run("Logs if no changes are required", func(t *testing.T) {
		// Arrange
		mockClient := new(MockDeploymentsClient)
		mockLogger := new(log.TestifyMock)

		deploymentConfig := acappsv1.Deployment("deployment", "namespace")

		k := &KubeClient{
			ctx:               context.Background(),
			deploymentsClient: mockClient,
			logger:            mockLogger,
		}

		dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
		mockClient.On("Apply", k.ctx, deploymentConfig, dryRunOpts).Once().Return(nil, nil)
		mockLogger.On("Info", "No changes required for deployment 'deployment'")

		// Act
		err := k.ApplyDeployment(deploymentConfig)

		// Assert
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)

		assert.Nil(t, err)
	})

	t.Run("Applies deployment spec if required and waits", func(t *testing.T) {
		// Arrange
		mockClient := new(MockDeploymentsClient)
		mockEventWaiter := new(MockEventWaiter)

		deploymentConfig := acappsv1.Deployment("deployment", "namespace")

		k := &KubeClient{
			ctx:               context.Background(),
			namespace:         "namespace",
			deploymentsClient: mockClient,
			waiter:            mockEventWaiter,
		}

		dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
		mockClient.On("Apply", k.ctx, deploymentConfig, dryRunOpts).Once().Return(&appsv1.Deployment{}, nil)

		actualRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli"}
		mockClient.On("Apply", k.ctx, deploymentConfig, actualRunOpts).Once().Return(nil, nil)

		waitErr := errors.New("waiter error")
		mockEventWaiter.On(
			"WaitFor",
			"deployment",
			"namespace",
			[]watch.EventType{watch.Added, watch.Modified},
			mockClient,
		).Once().Return(waitErr)

		// Act
		err := k.ApplyDeployment(deploymentConfig)

		// Assert
		mockClient.AssertExpectations(t)
		mockEventWaiter.AssertExpectations(t)

		assert.EqualError(t, err, waitErr.Error())
	})

	t.Run("Returns error is couldn't apply dry run", func(t *testing.T) {
		// Arrange
		mockClient := new(MockDeploymentsClient)
		mockLogger := new(log.TestifyMock)

		deploymentConfig := acappsv1.Deployment("deployment", "namespace")

		k := &KubeClient{
			ctx:               context.Background(),
			deploymentsClient: mockClient,
			logger:            mockLogger,
		}

		dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
		dryRunErr := errors.New("dry run err")
		mockClient.On("Apply", k.ctx, deploymentConfig, dryRunOpts).Once().Return(nil, dryRunErr)

		// Act
		err := k.ApplyDeployment(deploymentConfig)

		// Assert
		mockClient.AssertExpectations(t)
		mockLogger.AssertNotCalled(t, "Info")

		assert.EqualError(t, err, dryRunErr.Error())
	})

	t.Run("Returns error if couldn't apply real run", func(t *testing.T) {
		// Arrange
		mockClient := new(MockDeploymentsClient)
		mockEventWaiter := new(MockEventWaiter)

		deploymentConfig := acappsv1.Deployment("deployment", "namespace")

		k := &KubeClient{
			ctx:               context.Background(),
			namespace:         "namespace",
			deploymentsClient: mockClient,
			waiter:            mockEventWaiter,
		}

		dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
		mockClient.On("Apply", k.ctx, deploymentConfig, dryRunOpts).Once().Return(&appsv1.Deployment{}, nil)

		actualRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli"}
		actualRunErr := errors.New("actual run err")
		mockClient.On("Apply", k.ctx, deploymentConfig, actualRunOpts).Once().Return(nil, actualRunErr)

		// Act
		err := k.ApplyDeployment(deploymentConfig)

		// Assert
		mockClient.AssertExpectations(t)
		mockEventWaiter.AssertNotCalled(t, "WaitFor")

		assert.EqualError(t, err, actualRunErr.Error())
	})
}

type MockConfigBuilder struct {
	mock.Mock
}

func (m *MockConfigBuilder) GetKubeConfigPath() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockConfigBuilder) BuildConfig(clusterURL, kubeConfigPath string) (*rest.Config, error) {
	args := m.Called(clusterURL, kubeConfigPath)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*rest.Config), args.Error(1)
}

type MockKubeClientBuilder struct {
	mock.Mock
}

func (m *MockKubeClientBuilder) NewForConfig(config *rest.Config) (*kubernetes.Clientset, error) {
	args := m.Called(config)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*kubernetes.Clientset), args.Error(1)
}

type MockDeploymentsClient struct {
	mock.Mock
}

func (m *MockDeploymentsClient) Apply(
	ctx context.Context,
	deployment *acappsv1.DeploymentApplyConfiguration,
	opts metav1.ApplyOptions,
) (*appsv1.Deployment, error) {
	args := m.Called(ctx, deployment, opts)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*appsv1.Deployment), args.Error(1)
}

func (m *MockDeploymentsClient) Watch(
	ctx context.Context,
	options metav1.ListOptions,
) (watch.Interface, error) {
	args := m.Called(ctx, options)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(watch.Interface), args.Error(1)
}

type MockWatchable struct {
	mock.Mock
}

func (m *MockWatchable) Watch(ctx context.Context, options metav1.ListOptions) (watch.Interface, error) {
	args := m.Called(ctx, options)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(watch.Interface), args.Error(1)
}

type MockEventWaiter struct {
	mock.Mock
}

func (m *MockEventWaiter) WaitFor(resourceName, namespace string, eventTypes []watch.EventType, client watchable) error {
	args := m.Called(resourceName, namespace, eventTypes, client)
	return args.Error(0)
}
