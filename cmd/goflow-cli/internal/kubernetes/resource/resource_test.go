package resource

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
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

func TestNewNamespace(t *testing.T) {
	// Arrange
	mockClient := new(mockNamespaceClient)
	namespaceName := "test-namespace"

	config := acapiv1.Namespace(namespaceName)
	applyOptions := metav1.ApplyOptions{}
	getOptions := metav1.GetOptions{}

	mockNamespace := &apiv1.Namespace{}
	ctx := context.Background()

	// Act
	resource := NewNamespace(config, mockClient)

	t.Run("Initialises name and kind", func(t *testing.T) {
		// Assert
		assert.Equal(t, namespaceName, resource.name)
		assert.Equal(t, "namespace", resource.kind)
	})

	t.Run("Initialises applyFunc", func(t *testing.T) {
		mockClient.On("Apply", ctx, config, applyOptions).Return(mockNamespace, nil)
		result, err := resource.Apply(ctx, applyOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, mockNamespace, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Initialises getFunc", func(t *testing.T) {
		mockClient.On("Get", ctx, namespaceName, getOptions).Return(mockNamespace, nil)
		result, err := resource.Get(ctx, getOptions)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, mockNamespace, result)
		mockClient.AssertExpectations(t)
	})
}

type mockNamespaceClient struct {
	mock.Mock
}

func (m *mockNamespaceClient) Apply(ctx context.Context, config *acapiv1.NamespaceApplyConfiguration, opts metav1.ApplyOptions) (*apiv1.Namespace, error) {
	args := m.Called(ctx, config, opts)
	return args.Get(0).(*apiv1.Namespace), args.Error(1)
}

func (m *mockNamespaceClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiv1.Namespace, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(*apiv1.Namespace), args.Error(1)
}
