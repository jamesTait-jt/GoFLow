//go:build unit

package k8s

import (
	"errors"
	"testing"
	"time"

	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/watch"
)

func Test_NewDeploymentExecutor(t *testing.T) {
	t.Run("Initialises deployment executor", func(t *testing.T) {
		// Arrange
		logger := new(log.TestifyMock)

		// Act
		de := NewDeploymentExecutor(logger)

		// Assert
		assert.NotNil(t, de)
		assert.NotNil(t, de.op)
		assert.Equal(t, logger, de.logger)
	})
}

func Test_DeploymentExecutor_ApplyAndWait(t *testing.T) {
	t.Run("Successfully applies and waits for resource modification", func(t *testing.T) {
		// Arrange
		mockOp := new(mockOperator)
		mockLogger := new(log.TestifyMock)
		kubeResource := new(mockResource)
		applyTimeout := time.Second * 10

		d := &DeploymentExecutor{op: mockOp, logger: mockLogger}

		mockOp.On("Apply", kubeResource).Once().Return(true, nil)
		mockOp.On("WaitFor", kubeResource, []watch.EventType{watch.Added, watch.Modified}, applyTimeout).Once().Return(nil)

		kubeResource.On("Kind").Return("Deployment")
		kubeResource.On("Name").Return("example-deployment")

		mockLogger.On("Info", "Deploying Deployment 'example-deployment'").Once()
		mockLogger.On("Info", "'example-deployment' needs modification - applying changes").Once()
		mockLogger.On("Success", "'example-deployment' deployed successfully").Once()

		// Act
		err := d.ApplyAndWait(kubeResource, applyTimeout)

		// Assert
		assert.NoError(t, err)
		mockOp.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		kubeResource.AssertExpectations(t)
	})

	t.Run("Successfully applies with no modification needed", func(t *testing.T) {
		// Arrange
		mockOp := new(mockOperator)
		mockLogger := new(log.TestifyMock)
		kubeResource := new(mockResource)
		applyTimeout := time.Second * 10

		d := &DeploymentExecutor{op: mockOp, logger: mockLogger}

		mockOp.On("Apply", kubeResource).Once().Return(false, nil)

		kubeResource.On("Kind").Return("Deployment")
		kubeResource.On("Name").Return("example-deployment")

		mockLogger.On("Info", "Deploying Deployment 'example-deployment'").Once()
		mockLogger.On("Success", "'example-deployment' deployed successfully").Once()

		// Act
		err := d.ApplyAndWait(kubeResource, applyTimeout)

		// Assert
		assert.NoError(t, err)
		mockOp.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		kubeResource.AssertExpectations(t)
	})

	t.Run("Returns an error if Apply fails", func(t *testing.T) {
		// Arrange
		mockOp := new(mockOperator)
		mockLogger := new(log.TestifyMock)
		kubeResource := new(mockResource)
		applyTimeout := time.Second * 10

		d := &DeploymentExecutor{op: mockOp, logger: mockLogger}

		applyErr := errors.New("apply error")
		mockOp.On("Apply", kubeResource).Once().Return(false, applyErr)

		kubeResource.On("Kind").Return("Deployment")
		kubeResource.On("Name").Return("example-deployment")

		mockLogger.On("Info", "Deploying Deployment 'example-deployment'").Once()

		// Act
		err := d.ApplyAndWait(kubeResource, applyTimeout)

		// Assert
		assert.EqualError(t, err, applyErr.Error())
		mockOp.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		kubeResource.AssertExpectations(t)
	})

	t.Run("Returns an error if failed waiting for apply", func(t *testing.T) {
		// Arrange
		mockOp := new(mockOperator)
		mockLogger := new(log.TestifyMock)
		kubeResource := new(mockResource)
		applyTimeout := time.Second * 10

		d := &DeploymentExecutor{op: mockOp, logger: mockLogger}

		mockOp.On("Apply", kubeResource).Once().Return(true, nil)

		waitErr := errors.New("wait error")
		mockOp.On("WaitFor", kubeResource, []watch.EventType{watch.Added, watch.Modified}, applyTimeout).Once().Return(waitErr)

		kubeResource.On("Kind").Return("Deployment")
		kubeResource.On("Name").Return("example-deployment")

		mockLogger.On("Info", "Deploying Deployment 'example-deployment'").Once()
		mockLogger.On("Info", "'example-deployment' needs modification - applying changes").Once()

		// Act
		err := d.ApplyAndWait(kubeResource, applyTimeout)

		// Assert
		assert.EqualError(t, err, waitErr.Error())
		mockOp.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		kubeResource.AssertExpectations(t)
	})
}

func Test_DeploymentExecutor_DestroyAndWait(t *testing.T) {
	t.Run("Successfully deletes and waits for resource destruction", func(t *testing.T) {
		// Arrange
		mockOp := new(mockOperator)
		mockLogger := new(log.TestifyMock)
		kubeResource := new(mockResource)
		deleteTimeout := time.Second * 10

		d := &DeploymentExecutor{op: mockOp, logger: mockLogger}

		mockOp.On("Delete", kubeResource).Once().Return(true, nil)
		mockOp.On("WaitFor", kubeResource, []watch.EventType{watch.Deleted}, deleteTimeout).Once().Return(nil)

		kubeResource.On("Kind").Return("PersistentVolume")
		kubeResource.On("Name").Return("example-pv")

		mockLogger.On("Info", "Destroying PersistentVolume 'example-pv'").Once()
		mockLogger.On("Info", "'example-pv' needs destroying - waiting...").Once()
		mockLogger.On("Success", "'example-pv' destroyed successfully").Once()

		// Act
		err := d.DestroyAndWait(kubeResource, deleteTimeout)

		// Assert
		assert.NoError(t, err)
		mockOp.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		kubeResource.AssertExpectations(t)
	})

	t.Run("Successfully handles resource already deleted", func(t *testing.T) {
		// Arrange
		mockOp := new(mockOperator)
		mockLogger := new(log.TestifyMock)
		kubeResource := new(mockResource)
		deleteTimeout := time.Second * 10

		d := &DeploymentExecutor{op: mockOp, logger: mockLogger}

		mockOp.On("Delete", kubeResource).Once().Return(false, nil)

		kubeResource.On("Kind").Return("PersistentVolume")
		kubeResource.On("Name").Return("example-pv")

		mockLogger.On("Info", "Destroying PersistentVolume 'example-pv'").Once()
		mockLogger.On("Warn", "couldnt find 'example-pv'").Once()

		// Act
		err := d.DestroyAndWait(kubeResource, deleteTimeout)

		// Assert
		assert.NoError(t, err)
		mockOp.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		kubeResource.AssertExpectations(t)
	})

	t.Run("Returns an error if Delete operation fails", func(t *testing.T) {
		// Arrange
		mockOp := new(mockOperator)
		mockLogger := new(log.TestifyMock)
		kubeResource := new(mockResource)
		deleteTimeout := time.Second * 10

		d := &DeploymentExecutor{op: mockOp, logger: mockLogger}

		deleteErr := errors.New("delete error")
		mockOp.On("Delete", kubeResource).Once().Return(false, deleteErr)

		kubeResource.On("Kind").Return("PersistentVolume")
		kubeResource.On("Name").Return("example-pv")

		mockLogger.On("Info", "Destroying PersistentVolume 'example-pv'").Once()

		// Act
		err := d.DestroyAndWait(kubeResource, deleteTimeout)

		// Assert
		assert.EqualError(t, err, deleteErr.Error())
		mockOp.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		kubeResource.AssertExpectations(t)
	})

	t.Run("Returns an error if WaitFor failed waiting for apply", func(t *testing.T) {
		// Arrange
		mockOp := new(mockOperator)
		mockLogger := new(log.TestifyMock)
		kubeResource := new(mockResource)
		deleteTimeout := time.Second * 10

		d := &DeploymentExecutor{op: mockOp, logger: mockLogger}

		mockOp.On("Delete", kubeResource).Once().Return(true, nil)

		waitErr := errors.New("wait error")
		mockOp.On("WaitFor", kubeResource, []watch.EventType{watch.Deleted}, deleteTimeout).Once().Return(waitErr)

		kubeResource.On("Kind").Return("PersistentVolume")
		kubeResource.On("Name").Return("example-pv")

		mockLogger.On("Info", "Destroying PersistentVolume 'example-pv'").Once()
		mockLogger.On("Info", "'example-pv' needs destroying - waiting...").Once()

		// Act
		err := d.DestroyAndWait(kubeResource, deleteTimeout)

		// Assert
		assert.EqualError(t, err, waitErr.Error())
		mockOp.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		kubeResource.AssertExpectations(t)
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
