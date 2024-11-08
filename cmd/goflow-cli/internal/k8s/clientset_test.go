//go:build unit

package k8s

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Test_NewClientSet(t *testing.T) {
	t.Run("Returns clientset", func(t *testing.T) {
		// Arrange
		clusterURL := "cluster"
		namespace := "test-namespace"
		mockKubeConfBuilder := new(mockKubeConfigBuilder)
		mockClientsetBuilder := new(mockClientsetBuilder)

		kubeconfigPath := "config path"
		mockKubeConfBuilder.On("GetKubeConfigPath").Once().Return(kubeconfigPath, nil)

		kubeConfig := &rest.Config{}
		mockKubeConfBuilder.On("BuildConfig", clusterURL, kubeconfigPath).Once().Return(kubeConfig, nil)

		clientset := &kubernetes.Clientset{}
		mockClientsetBuilder.On("NewForConfig", kubeConfig).Once().Return(clientset, nil)

		expectedK8sClients := &Clients{
			clientset: clientset,
			namespace: namespace,
		}

		// Act
		cs, err := NewClientset(
			clusterURL,
			namespace,
			WithConfigBuilder(mockKubeConfBuilder),
			WithKubeClientBuilder(mockClientsetBuilder),
		)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, expectedK8sClients, cs)

		mockKubeConfBuilder.AssertExpectations(t)
		mockClientsetBuilder.AssertExpectations(t)
	})

	t.Run("Returns error if couldnt get kube conf path", func(t *testing.T) {
		// Arrange
		clusterURL := "cluster"
		namespace := "test-namespace"
		mockKubeConfBuilder := new(mockKubeConfigBuilder)
		mockClientsetBuilder := new(mockClientsetBuilder)

		getPathErr := errors.New("couldnt get path")
		mockKubeConfBuilder.On("GetKubeConfigPath").Once().Return("", getPathErr)

		// Act
		cs, err := NewClientset(
			clusterURL,
			namespace,
			WithConfigBuilder(mockKubeConfBuilder),
			WithKubeClientBuilder(mockClientsetBuilder),
		)

		// Assert
		assert.EqualError(t, err, getPathErr.Error())
		assert.Nil(t, cs)

		mockKubeConfBuilder.AssertExpectations(t)
		mockClientsetBuilder.AssertExpectations(t)
	})

	t.Run("Returns error if couldnt build kube conf", func(t *testing.T) {
		// Arrange
		clusterURL := "cluster"
		namespace := "test-namespace"
		mockKubeConfBuilder := new(mockKubeConfigBuilder)
		mockClientsetBuilder := new(mockClientsetBuilder)

		kubeconfigPath := "config path"
		mockKubeConfBuilder.On("GetKubeConfigPath").Once().Return(kubeconfigPath, nil)

		buildConfErr := errors.New("couldnt build conf")
		mockKubeConfBuilder.On("BuildConfig", clusterURL, kubeconfigPath).Once().Return(nil, buildConfErr)

		// Act
		cs, err := NewClientset(
			clusterURL,
			namespace,
			WithConfigBuilder(mockKubeConfBuilder),
			WithKubeClientBuilder(mockClientsetBuilder),
		)

		// Assert
		assert.EqualError(t, err, buildConfErr.Error())
		assert.Nil(t, cs)

		mockKubeConfBuilder.AssertExpectations(t)
		mockClientsetBuilder.AssertExpectations(t)
	})

	t.Run("Returns error if couldnt build clientset", func(t *testing.T) {
		// Arrange
		clusterURL := "cluster"
		namespace := "test-namespace"
		mockKubeConfBuilder := new(mockKubeConfigBuilder)
		mockClientsetBuilder := new(mockClientsetBuilder)

		kubeconfigPath := "config path"
		mockKubeConfBuilder.On("GetKubeConfigPath").Once().Return(kubeconfigPath, nil)

		kubeConfig := &rest.Config{}
		mockKubeConfBuilder.On("BuildConfig", clusterURL, kubeconfigPath).Once().Return(kubeConfig, nil)

		buildClientsetErr := errors.New("couldnt build clientset")
		mockClientsetBuilder.On("NewForConfig", kubeConfig).Once().Return(nil, buildClientsetErr)

		// Act
		cs, err := NewClientset(
			clusterURL,
			namespace,
			WithConfigBuilder(mockKubeConfBuilder),
			WithKubeClientBuilder(mockClientsetBuilder),
		)

		// Assert
		assert.EqualError(t, err, buildClientsetErr.Error())
		assert.Nil(t, cs)

		mockKubeConfBuilder.AssertExpectations(t)
		mockClientsetBuilder.AssertExpectations(t)
	})
}
