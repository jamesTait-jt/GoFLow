package kubernetes

import (
	"errors"
	"testing"

	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
