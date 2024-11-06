package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_DeploymentService_Deploy(t *testing.T) {
	t.Run("Successfully deploys all components", func(t *testing.T) {
		// Arrange
		mockManager := new(mockDeploymentManager)
		service := NewDeploymentService(mockManager)

		mockManager.On("DeployNamespace").Once().Return(nil)
		mockManager.On("DeployMessageBroker").Once().Return(nil)
		mockManager.On("DeployGRPCServer").Once().Return(nil)
		mockManager.On("DeployWorkerpools").Once().Return(nil)

		// Act
		err := service.Deploy()

		// Assert
		assert.NoError(t, err)
		mockManager.AssertExpectations(t)
	})

	t.Run("Returns an error if DeployNamespace fails", func(t *testing.T) {
		// Arrange
		mockManager := new(mockDeploymentManager)
		service := NewDeploymentService(mockManager)

		deployErr := errors.New("namespace deployment error")
		mockManager.On("DeployNamespace").Once().Return(deployErr)

		// Act
		err := service.Deploy()

		// Assert
		assert.EqualError(t, err, fmt.Sprintf("failed to deploy namespace: %s", deployErr))
		mockManager.AssertExpectations(t)
	})

	t.Run("Returns an error if DeployMessageBroker fails", func(t *testing.T) {
		// Arrange
		mockManager := new(mockDeploymentManager)
		service := NewDeploymentService(mockManager)

		mockManager.On("DeployNamespace").Once().Return(nil)

		deployErr := errors.New("message broker deployment error")
		mockManager.On("DeployMessageBroker").Once().Return(deployErr)

		// Act
		err := service.Deploy()

		// Assert
		assert.EqualError(t, err, fmt.Sprintf("failed to deploy message broker: %s", deployErr))
		mockManager.AssertExpectations(t)
	})

	t.Run("Returns an error if DeployGRPCServer fails", func(t *testing.T) {
		// Arrange
		mockManager := new(mockDeploymentManager)
		service := NewDeploymentService(mockManager)

		mockManager.On("DeployNamespace").Once().Return(nil)
		mockManager.On("DeployMessageBroker").Once().Return(nil)

		deployErr := errors.New("gRPC server deployment error")
		mockManager.On("DeployGRPCServer").Once().Return(deployErr)

		// Act
		err := service.Deploy()

		// Assert
		assert.EqualError(t, err, fmt.Sprintf("failed to deploy gRPC server: %s", deployErr))
		mockManager.AssertExpectations(t)
	})

	t.Run("Returns an error if DeployWorkerpools fails", func(t *testing.T) {
		// Arrange
		mockManager := new(mockDeploymentManager)
		service := NewDeploymentService(mockManager)

		mockManager.On("DeployNamespace").Once().Return(nil)
		mockManager.On("DeployMessageBroker").Once().Return(nil)
		mockManager.On("DeployGRPCServer").Once().Return(nil)

		deployErr := errors.New("worker pools deployment error")
		mockManager.On("DeployWorkerpools").Once().Return(deployErr)

		// Act
		err := service.Deploy()

		// Assert
		assert.EqualError(t, err, fmt.Sprintf("failed to deploy worker pools: %s", deployErr))
		mockManager.AssertExpectations(t)
	})
}

func Test_DeploymentService_Destroy(t *testing.T) {
	t.Run("Successfully destroys all components", func(t *testing.T) {
		// Arrange
		mockManager := new(mockDeploymentManager)
		service := NewDeploymentService(mockManager)

		mockManager.On("DestroyAll").Once().Return(nil)

		// Act
		err := service.Destroy()

		// Assert
		assert.NoError(t, err)
		mockManager.AssertExpectations(t)
	})

	t.Run("Returns an error if DestroyAll fails", func(t *testing.T) {
		// Arrange
		mockManager := new(mockDeploymentManager)
		service := NewDeploymentService(mockManager)

		destroyErr := errors.New("destroy error")
		mockManager.On("DestroyAll").Once().Return(destroyErr)

		// Act
		err := service.Destroy()

		// Assert
		assert.EqualError(t, err, destroyErr.Error())
		mockManager.AssertExpectations(t)
	})
}

type mockDeploymentManager struct {
	mock.Mock
}

func (m *mockDeploymentManager) DeployNamespace() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockDeploymentManager) DeployMessageBroker() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockDeploymentManager) DeployGRPCServer() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockDeploymentManager) DeployWorkerpools() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockDeploymentManager) DestroyAll() error {
	args := m.Called()
	return args.Error(0)
}
