package resource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_NewNamespaceApplier(t *testing.T) {
	t.Run("Initialises namespace applier correctly", func(t *testing.T) {
		// Arrange
		clientset := fake.NewClientset()

		// Act
		namespaceApplier := NewNamespaceApplier(clientset)

		// Assert
		assert.NotNil(t, namespaceApplier)
		assert.NotNil(t, namespaceApplier.client)
		assert.IsType(t, &ObjectSpeccer{}, namespaceApplier.speccer)
	})
}

func Test_NewDeploymentApplier(t *testing.T) {
	t.Run("Initialises deployment applier correctly", func(t *testing.T) {
		// Arrange
		clientset := fake.NewClientset()

		// Act
		deploymentApplier := NewDeploymentApplier(clientset, "namespace")

		// Assert
		assert.NotNil(t, deploymentApplier)
		assert.NotNil(t, deploymentApplier.client)
		assert.IsType(t, &ObjectSpeccer{}, deploymentApplier.speccer)
	})
}

func Test_NewServiceApplier(t *testing.T) {
	t.Run("Initialises service applier correctly", func(t *testing.T) {
		// Arrange
		clientset := fake.NewClientset()

		// Act
		serviceApplier := NewServiceApplier(clientset, "namespace")

		// Assert
		assert.NotNil(t, serviceApplier)
		assert.NotNil(t, serviceApplier.client)
		assert.IsType(t, &ObjectSpeccer{}, serviceApplier.speccer)
	})
}

func Test_Applier_Apply(t *testing.T) {
	t.Run("Returns error if failed to get current resource", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockClient := new(mockGetApplier[*mockAppliable, *mockRuntimeObject])

		appliable := &mockAppliable{}

		a := &Applier[*mockAppliable, *mockRuntimeObject]{
			client: mockClient,
		}

		name := "appliableName"
		appliable.On("GetName").Once().Return(&name)

		mockClient.On("Get")

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

	// t.Run("Returns false if no changes are required", func(t *testing.T) {
	// 	// Arrange
	// 	mockClient := new(mockApplyWatchable[mockHasName, any])
	// 	mockLogger := new(log.TestifyMock)
	// 	mockWaiter := new(mockEventWaiter)

	// 	name := "testName"
	// 	nameable := mockHasName{
	// 		name: name,
	// 	}

	// 	a := &Applier[mockHasName, any]{
	// 		ctx:    context.Background(),
	// 		client: mockClient,
	// 		logger: mockLogger,
	// 		waiter: mockWaiter,
	// 	}

	// 	dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
	// 	mockClient.On("Apply", a.ctx, nameable, dryRunOpts).Once().Return(nil, nil)
	// 	mockLogger.On("Info", "No changes required for 'testName'")

	// 	// Act
	// 	err := a.Apply(nameable, "namespace")

	// 	// Assert
	// 	mockClient.AssertExpectations(t)
	// 	mockLogger.AssertExpectations(t)

	// 	assert.Nil(t, err)
	// })

	// 	t.Run("Applies spec if required and waits", func(t *testing.T) {
	// 		// Arrange
	// 		mockClient := new(mockApplyWatchable[mockHasName, any])
	// 		mockWaiter := new(mockEventWaiter)

	// 		name := "testName"
	// 		nameable := mockHasName{
	// 			name: name,
	// 		}

	// 		namespace := "testNamespace"

	// 		a := &Applier[mockHasName, any]{
	// 			ctx:    context.Background(),
	// 			client: mockClient,
	// 			waiter: mockWaiter,
	// 		}

	// 		dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
	// 		mockClient.On("Apply", a.ctx, nameable, dryRunOpts).Once().Return(new(any), nil)

	// 		actualRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli"}
	// 		mockClient.On("Apply", a.ctx, nameable, actualRunOpts).Once().Return(nil, nil)

	// 		waitErr := errors.New("waiter error")
	// 		mockWaiter.On(
	// 			"WaitFor",
	// 			name,
	// 			namespace,
	// 			[]watch.EventType{watch.Added, watch.Modified},
	// 			mockClient,
	// 		).Once().Return(waitErr)

	// 		// Act
	// 		err := a.Apply(nameable, namespace)

	// 		// Assert
	// 		mockClient.AssertExpectations(t)
	// 		mockWaiter.AssertExpectations(t)

	// 		assert.EqualError(t, err, waitErr.Error())
	// 	})

	// 	t.Run("Returns error if couldn't apply dry run", func(t *testing.T) {
	// 		mockClient := new(mockApplyWatchable[mockHasName, any])
	// 		mockWaiter := new(mockEventWaiter)

	// 		name := "testName"
	// 		nameable := mockHasName{
	// 			name: name,
	// 		}

	// 		a := &Applier[mockHasName, any]{
	// 			ctx:    context.Background(),
	// 			client: mockClient,
	// 			waiter: mockWaiter,
	// 		}

	// 		dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
	// 		dryRunErr := errors.New("dry run err")
	// 		mockClient.On("Apply", a.ctx, nameable, dryRunOpts).Once().Return(nil, dryRunErr)

	// 		// Act
	// 		err := a.Apply(nameable, "")

	// 		// Assert
	// 		mockClient.AssertExpectations(t)

	// 		assert.EqualError(t, err, dryRunErr.Error())
	// 	})

	// 	t.Run("Returns error if couldn't apply real run", func(t *testing.T) {
	// 		// Arrange
	// 		mockClient := new(mockApplyWatchable[mockHasName, any])
	// 		mockWaiter := new(mockEventWaiter)

	// 		name := "testName"
	// 		nameable := mockHasName{
	// 			name: name,
	// 		}

	// 		a := &Applier[mockHasName, any]{
	// 			ctx:    context.Background(),
	// 			client: mockClient,
	// 			waiter: mockWaiter,
	// 		}

	// 		dryRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}
	// 		mockClient.On("Apply", a.ctx, nameable, dryRunOpts).Once().Return(new(any), nil)

	// 		actualRunOpts := metav1.ApplyOptions{FieldManager: "goflow-cli"}
	// 		actualRunErr := errors.New("actual run err")
	// 		mockClient.On("Apply", a.ctx, nameable, actualRunOpts).Once().Return(nil, actualRunErr)

	// 		// Act
	// 		err := a.Apply(nameable, "")

	// 		// Assert
	// 		mockClient.AssertExpectations(t)
	// 		mockWaiter.AssertNotCalled(t, "WaitFor")

	//		assert.EqualError(t, err, actualRunErr.Error())
	//	})
}

type mockRuntimeObject struct {
	mock.Mock
}

func (m *mockRuntimeObject) GetObjectKind() schema.ObjectKind {
	args := m.Called()
	return args.Get(0).(schema.ObjectKind)
}

func (m *mockRuntimeObject) DeepCopyObject() runtime.Object {
	args := m.Called()
	return args.Get(0).(runtime.Object)
}

type mockAppliable struct {
	mock.Mock
}

func (m *mockAppliable) GetName() *string {
	args := m.Called()
	return args.Get(0).(*string)
}

type mockGetApplier[C Appliable, R runtime.Object] struct {
	mock.Mock
}

func (m *mockGetApplier[C, R]) Get(ctx context.Context, name string, opts metav1.GetOptions) (R, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(R), args.Error(1)
}

func (m *mockGetApplier[C, R]) Apply(ctx context.Context, applyConfig C, opts metav1.ApplyOptions) (R, error) {
	args := m.Called(ctx, applyConfig, opts)
	return args.Get(0).(R), args.Error(1)
}
