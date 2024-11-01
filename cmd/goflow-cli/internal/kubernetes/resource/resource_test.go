package resource

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"

	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
)

// nolint:dupl // it is easier to understand what these tests are doing with the duplication
func Test_Resource_Apply(t *testing.T) {
	t.Run("Calls applyFunc", func(t *testing.T) {
		// Arrange
		called := false

		var receivedCtx context.Context

		var receivedOpts metav1.ApplyOptions

		returnedRuntimeObject := &runtime.Unknown{}
		returnedErr := errors.New("error")

		r := Resource{
			applyFunc: func(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
				called = true
				receivedCtx = ctx
				receivedOpts = opts

				return returnedRuntimeObject, returnedErr
			},
		}

		sentCtx := context.Background()
		sentOpts := metav1.ApplyOptions{}

		// Act
		actualRuntimeObject, actualErr := r.Apply(sentCtx, sentOpts)

		// Assert
		assert.True(t, called)
		assert.Equal(t, sentCtx, receivedCtx)
		assert.Equal(t, sentOpts, receivedOpts)
		assert.Equal(t, returnedRuntimeObject, actualRuntimeObject)
		assert.EqualError(t, returnedErr, actualErr.Error())
	})
}

// nolint:dupl // it is easier to understand what these tests are doing with the duplication
func Test_Resource_Get(t *testing.T) {
	t.Run("Calls getFunc", func(t *testing.T) {
		// Arrange
		called := false

		var receivedCtx context.Context

		var receivedOpts metav1.GetOptions

		returnedRuntimeObject := &runtime.Unknown{}
		returnedErr := errors.New("error")

		r := Resource{
			getFunc: func(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
				called = true
				receivedCtx = ctx
				receivedOpts = opts

				return returnedRuntimeObject, returnedErr
			},
		}

		sentCtx := context.Background()
		sentOpts := metav1.GetOptions{}

		// Act
		actualRuntimeObject, actualErr := r.Get(sentCtx, sentOpts)

		// Assert
		assert.True(t, called)
		assert.Equal(t, sentCtx, receivedCtx)
		assert.Equal(t, sentOpts, receivedOpts)
		assert.Equal(t, returnedRuntimeObject, actualRuntimeObject)
		assert.EqualError(t, returnedErr, actualErr.Error())
	})
}

// nolint:dupl // it is easier to understand what these tests are doing with the duplication
func Test_NewNamespace(t *testing.T) {
	// Arrange
	mockClient := new(mockNamespaceClient)
	namespaceName := "test-namespace"

	config := acapiv1.Namespace(namespaceName)
	applyOptions := metav1.ApplyOptions{}
	getOptions := metav1.GetOptions{}

	returnedNamespace := &apiv1.Namespace{}
	ctx := context.Background()

	// Act
	resource := NewNamespace(config, mockClient)

	t.Run("Initialises name and kind", func(t *testing.T) {
		// Assert
		assert.Equal(t, namespaceName, resource.name)
		assert.Equal(t, "namespace", resource.kind)
	})

	t.Run("Initialises applyFunc", func(t *testing.T) {
		mockClient.On("Apply", ctx, config, applyOptions).Return(returnedNamespace, nil)
		result, err := resource.Apply(ctx, applyOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, returnedNamespace, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Initialises getFunc", func(t *testing.T) {
		mockClient.On("Get", ctx, namespaceName, getOptions).Return(returnedNamespace, nil)
		result, err := resource.Get(ctx, getOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, returnedNamespace, result)
		mockClient.AssertExpectations(t)
	})
}

// nolint:dupl // it is easier to understand what these tests are doing with the duplication
func Test_NewDeploymemt(t *testing.T) {
	// Arrange
	mockClient := new(mockDeploymentClient)
	deploymentName := "test-deployment"
	deploymentNamespace := "test-namespace"

	config := acappsv1.Deployment(deploymentName, deploymentNamespace)
	applyOptions := metav1.ApplyOptions{}
	getOptions := metav1.GetOptions{}

	returnedDeployment := &appsv1.Deployment{}
	ctx := context.Background()

	// Act
	resource := NewDeployment(config, mockClient)

	t.Run("Initialises name and kind", func(t *testing.T) {
		// Assert
		assert.Equal(t, deploymentName, resource.name)
		assert.Equal(t, "deployment", resource.kind)
	})

	t.Run("Initialises applyFunc", func(t *testing.T) {
		mockClient.On("Apply", ctx, config, applyOptions).Return(returnedDeployment, nil)
		result, err := resource.Apply(ctx, applyOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, returnedDeployment, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Initialises getFunc", func(t *testing.T) {
		mockClient.On("Get", ctx, deploymentName, getOptions).Return(returnedDeployment, nil)
		result, err := resource.Get(ctx, getOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, returnedDeployment, result)
		mockClient.AssertExpectations(t)
	})
}

// nolint:dupl // it is easier to understand what these tests are doing with the duplication
func Test_NewService(t *testing.T) {
	// Arrange
	mockClient := new(mockServiceClient)
	serviceName := "test-service"
	serviceNamespace := "test-namespace"

	config := acapiv1.Service(serviceName, serviceNamespace)
	applyOptions := metav1.ApplyOptions{}
	getOptions := metav1.GetOptions{}

	returnedService := &apiv1.Service{}
	ctx := context.Background()

	// Act
	resource := NewService(config, mockClient)

	t.Run("Initialises name and kind", func(t *testing.T) {
		// Assert
		assert.Equal(t, serviceName, resource.name)
		assert.Equal(t, "service", resource.kind)
	})

	t.Run("Initialises applyFunc", func(t *testing.T) {
		mockClient.On("Apply", ctx, config, applyOptions).Return(returnedService, nil)
		result, err := resource.Apply(ctx, applyOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, returnedService, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Initialises getFunc", func(t *testing.T) {
		mockClient.On("Get", ctx, serviceName, getOptions).Return(returnedService, nil)
		result, err := resource.Get(ctx, getOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, returnedService, result)
		mockClient.AssertExpectations(t)
	})
}

// nolint:dupl // it is easier to understand what these tests are doing with the duplication
func Test_NewPersistentVolume(t *testing.T) {
	// Arrange
	mockClient := new(mockPersistentVolumeClient)
	pvName := "test-persistent-volume"

	config := acapiv1.PersistentVolume(pvName)
	applyOptions := metav1.ApplyOptions{}
	getOptions := metav1.GetOptions{}

	returnedPV := &apiv1.PersistentVolume{}
	ctx := context.Background()

	// Act
	resource := NewPersistentVolume(config, mockClient)

	t.Run("Initialises name and kind", func(t *testing.T) {
		// Assert
		assert.Equal(t, pvName, resource.name)
		assert.Equal(t, "pv", resource.kind)
	})

	t.Run("Initialises applyFunc", func(t *testing.T) {
		mockClient.On("Apply", ctx, config, applyOptions).Return(returnedPV, nil)
		result, err := resource.Apply(ctx, applyOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, returnedPV, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Initialises getFunc", func(t *testing.T) {
		mockClient.On("Get", ctx, pvName, getOptions).Return(returnedPV, nil)
		result, err := resource.Get(ctx, getOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, returnedPV, result)
		mockClient.AssertExpectations(t)
	})
}

// nolint:dupl // it is easier to understand what these tests are doing with the duplication
func Test_NewPersistentVolumeClaim(t *testing.T) {
	// Arrange
	mockClient := new(mockPersistentVolumeClaimClient)
	pvcName := "test-persistent-volume-claim"
	pvcNamspace := "test-namespace"

	config := acapiv1.PersistentVolumeClaim(pvcName, pvcNamspace)
	applyOptions := metav1.ApplyOptions{}
	getOptions := metav1.GetOptions{}

	returnedClaim := &apiv1.PersistentVolumeClaim{}
	ctx := context.Background()

	// Act
	resource := NewPersistentVolumeClaim(config, mockClient)

	t.Run("Initialises name and kind", func(t *testing.T) {
		// Assert
		assert.Equal(t, pvcName, resource.name)
		assert.Equal(t, "pvc", resource.kind)
	})

	t.Run("Initialises applyFunc", func(t *testing.T) {
		mockClient.On("Apply", ctx, config, applyOptions).Return(returnedClaim, nil)
		result, err := resource.Apply(ctx, applyOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, returnedClaim, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Initialises getFunc", func(t *testing.T) {
		mockClient.On("Get", ctx, pvcName, getOptions).Return(returnedClaim, nil)
		result, err := resource.Get(ctx, getOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, returnedClaim, result)
		mockClient.AssertExpectations(t)
	})
}

type mockNamespaceClient struct {
	mock.Mock
}

func (m *mockNamespaceClient) Apply(
	ctx context.Context, config *acapiv1.NamespaceApplyConfiguration, opts metav1.ApplyOptions,
) (*apiv1.Namespace, error) {
	args := m.Called(ctx, config, opts)
	return args.Get(0).(*apiv1.Namespace), args.Error(1)
}

func (m *mockNamespaceClient) Get(
	ctx context.Context, name string, opts metav1.GetOptions,
) (*apiv1.Namespace, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(*apiv1.Namespace), args.Error(1)
}

type mockDeploymentClient struct {
	mock.Mock
}

func (m *mockDeploymentClient) Apply(
	ctx context.Context, config *acappsv1.DeploymentApplyConfiguration, opts metav1.ApplyOptions,
) (*appsv1.Deployment, error) {
	args := m.Called(ctx, config, opts)
	return args.Get(0).(*appsv1.Deployment), args.Error(1)
}

func (m *mockDeploymentClient) Get(
	ctx context.Context, name string, opts metav1.GetOptions,
) (*appsv1.Deployment, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(*appsv1.Deployment), args.Error(1)
}

type mockServiceClient struct {
	mock.Mock
}

func (m *mockServiceClient) Apply(
	ctx context.Context, config *acapiv1.ServiceApplyConfiguration, opts metav1.ApplyOptions,
) (*apiv1.Service, error) {
	args := m.Called(ctx, config, opts)
	return args.Get(0).(*apiv1.Service), args.Error(1)
}

func (m *mockServiceClient) Get(
	ctx context.Context, name string, opts metav1.GetOptions,
) (*apiv1.Service, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(*apiv1.Service), args.Error(1)
}

type mockPersistentVolumeClient struct {
	mock.Mock
}

func (m *mockPersistentVolumeClient) Apply(
	ctx context.Context, config *acapiv1.PersistentVolumeApplyConfiguration, opts metav1.ApplyOptions,
) (*apiv1.PersistentVolume, error) {
	args := m.Called(ctx, config, opts)
	return args.Get(0).(*apiv1.PersistentVolume), args.Error(1)
}

func (m *mockPersistentVolumeClient) Get(
	ctx context.Context, name string, opts metav1.GetOptions,
) (*apiv1.PersistentVolume, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(*apiv1.PersistentVolume), args.Error(1)
}

type mockPersistentVolumeClaimClient struct {
	mock.Mock
}

func (m *mockPersistentVolumeClaimClient) Apply(
	ctx context.Context, config *acapiv1.PersistentVolumeClaimApplyConfiguration, opts metav1.ApplyOptions,
) (*apiv1.PersistentVolumeClaim, error) {
	args := m.Called(ctx, config, opts)
	return args.Get(0).(*apiv1.PersistentVolumeClaim), args.Error(1)
}

func (m *mockPersistentVolumeClaimClient) Get(
	ctx context.Context, name string, opts metav1.GetOptions,
) (*apiv1.PersistentVolumeClaim, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(*apiv1.PersistentVolumeClaim), args.Error(1)
}
