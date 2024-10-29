package resource

import (
	"context"
	"errors"
	"testing"

	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_NewNamespaceApplier(t *testing.T) {
	t.Run("Initialises namespace applier correctly", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		clientset := fake.NewClientset()
		logger := &log.TestifyMock{}
		mockWaiter := new(mockEventWaiter)

		// Act
		namespaceApplier := NewNamespaceApplier(ctx, clientset, logger, mockWaiter)

		// Assert
		assert.NotNil(t, namespaceApplier)
		assert.Equal(t, ctx, namespaceApplier.ctx)
		assert.NotNil(t, namespaceApplier.client)
		assert.Equal(t, logger, namespaceApplier.logger)
		assert.Equal(t, mockWaiter, namespaceApplier.waiter)
	})
}

func Test_NewDeploymentApplier(t *testing.T) {
	t.Run("Initialises deployment applier correctly", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		clientset := fake.NewClientset()
		logger := &log.TestifyMock{}
		mockWaiter := new(mockEventWaiter)
		namespace := "namespace"

		// Act
		deploymentApplier := NewDeploymentApplier(ctx, clientset, namespace, logger, mockWaiter)

		// Assert
		assert.NotNil(t, deploymentApplier)
		assert.Equal(t, ctx, deploymentApplier.ctx)
		assert.NotNil(t, deploymentApplier.client)
		assert.Equal(t, logger, deploymentApplier.logger)
		assert.Equal(t, mockWaiter, deploymentApplier.waiter)
	})
}

func Test_NewServiceApplier(t *testing.T) {
	t.Run("Initialises service applier correctly", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		clientset := fake.NewClientset()
		logger := &log.TestifyMock{}
		mockWaiter := new(mockEventWaiter)
		namespace := "namespace"

		// Act
		serviceApplier := NewServiceApplier(ctx, clientset, namespace, logger, mockWaiter)

		// Assert
		assert.NotNil(t, serviceApplier)
		assert.Equal(t, ctx, serviceApplier.ctx)
		assert.NotNil(t, serviceApplier.client)
		assert.Equal(t, logger, serviceApplier.logger)
		assert.Equal(t, mockWaiter, serviceApplier.waiter)
	})
}

func Test_Applier_Apply(t *testing.T) {
	t.Run("Logs if no changes are required", func(t *testing.T) {
		// Arrange
		mockClient := new(mockApplyWatchable[mockHasName, any])
		mockLogger := new(log.TestifyMock)
		mockWaiter := new(mockEventWaiter)

		name := "testName"
		nameable := mockHasName{
			name: name,
		}

		a := &Applier[mockHasName, any]{
			ctx:    context.Background(),
			client: mockClient,
			logger: mockLogger,
			waiter: mockWaiter,
		}

		dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
		mockClient.On("Apply", a.ctx, nameable, dryRunOpts).Once().Return(nil, nil)
		mockLogger.On("Info", "No changes required for 'testName'")

		// Act
		err := a.Apply(nameable, "namespace")

		// Assert
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)

		assert.Nil(t, err)
	})

	t.Run("Applies spec if required and waits", func(t *testing.T) {
		// Arrange
		mockClient := new(mockApplyWatchable[mockHasName, any])
		mockWaiter := new(mockEventWaiter)

		name := "testName"
		nameable := mockHasName{
			name: name,
		}

		namespace := "testNamespace"

		a := &Applier[mockHasName, any]{
			ctx:    context.Background(),
			client: mockClient,
			waiter: mockWaiter,
		}

		dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
		mockClient.On("Apply", a.ctx, nameable, dryRunOpts).Once().Return(new(any), nil)

		actualRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli"}
		mockClient.On("Apply", a.ctx, nameable, actualRunOpts).Once().Return(nil, nil)

		waitErr := errors.New("waiter error")
		mockWaiter.On(
			"WaitFor",
			name,
			namespace,
			[]watch.EventType{watch.Added, watch.Modified},
			mockClient,
		).Once().Return(waitErr)

		// Act
		err := a.Apply(nameable, namespace)

		// Assert
		mockClient.AssertExpectations(t)
		mockWaiter.AssertExpectations(t)

		assert.EqualError(t, err, waitErr.Error())
	})

	t.Run("Returns error if couldn't apply dry run", func(t *testing.T) {
		mockClient := new(mockApplyWatchable[mockHasName, any])
		mockWaiter := new(mockEventWaiter)

		name := "testName"
		nameable := mockHasName{
			name: name,
		}

		a := &Applier[mockHasName, any]{
			ctx:    context.Background(),
			client: mockClient,
			waiter: mockWaiter,
		}

		dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
		dryRunErr := errors.New("dry run err")
		mockClient.On("Apply", a.ctx, nameable, dryRunOpts).Once().Return(nil, dryRunErr)

		// Act
		err := a.Apply(nameable, "")

		// Assert
		mockClient.AssertExpectations(t)

		assert.EqualError(t, err, dryRunErr.Error())
	})

	t.Run("Returns error if couldn't apply real run", func(t *testing.T) {
		// Arrange
		mockClient := new(mockApplyWatchable[mockHasName, any])
		mockWaiter := new(mockEventWaiter)

		name := "testName"
		nameable := mockHasName{
			name: name,
		}

		a := &Applier[mockHasName, any]{
			ctx:    context.Background(),
			client: mockClient,
			waiter: mockWaiter,
		}

		dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
		mockClient.On("Apply", a.ctx, nameable, dryRunOpts).Once().Return(new(any), nil)

		actualRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli"}
		actualRunErr := errors.New("actual run err")
		mockClient.On("Apply", a.ctx, nameable, actualRunOpts).Once().Return(nil, actualRunErr)

		// Act
		err := a.Apply(nameable, "")

		// Assert
		mockClient.AssertExpectations(t)
		mockWaiter.AssertNotCalled(t, "WaitFor")

		assert.EqualError(t, err, actualRunErr.Error())
	})
}

type mockHasName struct {
	name string
}

func (m mockHasName) GetName() *string {
	return &m.name
}

type mockApplyWatchable[C hasName, R any] struct {
	mockWatchable
	mock.Mock
}

func (m *mockApplyWatchable[C, R]) Apply(
	ctx context.Context,
	configuration C,
	opts metav1.ApplyOptions,
) (*R, error) {
	args := m.Called(ctx, configuration, opts)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*R), args.Error(1)
}

type mockEventWaiter struct {
	mock.Mock
}

func (m *mockEventWaiter) WaitFor(resourceName, namespace string, eventTypes []watch.EventType, client Watchable) error {
	args := m.Called(resourceName, namespace, eventTypes, client)
	return args.Error(0)
}
