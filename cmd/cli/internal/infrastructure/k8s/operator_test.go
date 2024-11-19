//go:build unit

package k8s

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
)

func Test_NewOperator(t *testing.T) {
	t.Run("Initialises Operator", func(t *testing.T) {
		// Arrange
		// Act
		kube := NewOperator()

		// Assert
		assert.NotNil(t, kube.ctx)
		assert.NotNil(t, kube.speccer)
	})
}

func Test_Operator_Apply(t *testing.T) {
	t.Run("Applies the resource and returns true if the resource doesn't exist", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockResource := new(mockResource)
		mockSpeccer := new(mockSpeccer)

		o := &Operator{
			ctx:     ctx,
			speccer: mockSpeccer,
		}

		currResource := &runtime.Unknown{}
		notFoundErr := k8serr.NewNotFound(schema.GroupResource{}, "")
		mockResource.On("Get", ctx, metav1.GetOptions{}).Once().Return(currResource, notFoundErr)

		proposedResource := &runtime.Unknown{}
		mockResource.On(
			"Apply",
			ctx,
			metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}},
		).Once().Return(proposedResource, nil)

		currSpec := "CURRENTSPEC"
		mockSpeccer.On("Spec", currResource).Once().Return(currSpec, nil)

		proposedSpec := "PROPOSEDSPEC"
		mockSpeccer.On("Spec", proposedResource).Once().Return(proposedSpec, nil)

		mockResource.On(
			"Apply",
			ctx,
			metav1.ApplyOptions{FieldManager: "goflow-cli"},
		).Once().Return(nil, nil)

		// Act
		neededModification, err := o.Apply(mockResource)

		// Assert
		assert.Nil(t, err)
		assert.True(t, neededModification)

		mockResource.AssertExpectations(t)
		mockSpeccer.AssertExpectations(t)
	})

	t.Run("Applies the resource and returns true if the specs are different", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockResource := new(mockResource)
		mockSpeccer := new(mockSpeccer)

		o := &Operator{
			ctx:     ctx,
			speccer: mockSpeccer,
		}

		currResource := &runtime.Unknown{}
		mockResource.On("Get", ctx, metav1.GetOptions{}).Once().Return(currResource, nil)

		proposedResource := &runtime.Unknown{}
		mockResource.On(
			"Apply",
			ctx,
			metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}},
		).Once().Return(proposedResource, nil)

		currSpec := "CURRENTSPEC"
		mockSpeccer.On("Spec", currResource).Once().Return(currSpec, nil)

		proposedSpec := "PROPOSEDSPEC"
		mockSpeccer.On("Spec", proposedResource).Once().Return(proposedSpec, nil)

		mockResource.On(
			"Apply",
			ctx,
			metav1.ApplyOptions{FieldManager: "goflow-cli"},
		).Once().Return(nil, nil)

		// Act
		neededModification, err := o.Apply(mockResource)

		// Assert
		assert.Nil(t, err)
		assert.True(t, neededModification)

		mockResource.AssertExpectations(t)
		mockSpeccer.AssertExpectations(t)
	})

	t.Run("Does not apply the resource and returns false if the specs are the same", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockResource := new(mockResource)
		mockSpeccer := new(mockSpeccer)

		o := &Operator{
			ctx:     ctx,
			speccer: mockSpeccer,
		}

		currResource := &runtime.Unknown{}
		mockResource.On("Get", ctx, metav1.GetOptions{}).Once().Return(currResource, nil)

		proposedResource := &runtime.Unknown{}
		mockResource.On("Apply", ctx, metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}).
			Once().
			Return(proposedResource, nil)

		currSpec := "CURRENTSPEC"
		mockSpeccer.On("Spec", currResource).Once().Return(currSpec, nil)
		mockSpeccer.On("Spec", proposedResource).Once().Return(currSpec, nil)

		// Act
		neededModification, err := o.Apply(mockResource)

		// Assert
		assert.Nil(t, err)
		assert.False(t, neededModification)

		mockResource.AssertExpectations(t)
		mockSpeccer.AssertExpectations(t)
	})

	t.Run("Returns error if could not get current resource", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockResource := new(mockResource)
		mockSpeccer := new(mockSpeccer)

		o := &Operator{
			ctx:     ctx,
			speccer: mockSpeccer,
		}

		currResource := &runtime.Unknown{}
		getErr := errors.New("couldnt get")
		mockResource.On("Get", ctx, metav1.GetOptions{}).Once().Return(currResource, getErr)

		// Act
		neededModification, err := o.Apply(mockResource)

		// Assert
		assert.EqualError(t, err, getErr.Error())
		assert.False(t, neededModification)

		mockResource.AssertExpectations(t)
		mockSpeccer.AssertExpectations(t)
	})

	t.Run("Returns error if could not dry run apply", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockResource := new(mockResource)
		mockSpeccer := new(mockSpeccer)

		o := &Operator{
			ctx:     ctx,
			speccer: mockSpeccer,
		}

		currResource := &runtime.Unknown{}
		mockResource.On("Get", ctx, metav1.GetOptions{}).Once().Return(currResource, nil)

		applyErr := errors.New("could not dry run")
		mockResource.On("Apply", ctx, metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}).
			Once().
			Return(nil, applyErr)

		// Act
		neededModification, err := o.Apply(mockResource)

		// Assert
		assert.EqualError(t, err, applyErr.Error())
		assert.False(t, neededModification)

		mockResource.AssertExpectations(t)
		mockSpeccer.AssertExpectations(t)
	})

	t.Run("Returns error if could not get current spec of resource", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockResource := new(mockResource)
		mockSpeccer := new(mockSpeccer)

		o := &Operator{
			ctx:     ctx,
			speccer: mockSpeccer,
		}

		currResource := &runtime.Unknown{}
		mockResource.On("Get", ctx, metav1.GetOptions{}).Once().Return(currResource, nil)

		proposedResource := &runtime.Unknown{}
		mockResource.On("Apply", ctx, metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}).
			Once().
			Return(proposedResource, nil)

		specErr := errors.New("couldnt spec first")
		mockSpeccer.On("Spec", currResource).Once().Return(nil, specErr)

		// Act
		neededModification, err := o.Apply(mockResource)

		// Assert
		assert.EqualError(t, err, specErr.Error())
		assert.False(t, neededModification)

		mockResource.AssertExpectations(t)
		mockSpeccer.AssertExpectations(t)
	})

	t.Run("Returns error if could not get current spec of proposed resource", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockResource := new(mockResource)
		mockSpeccer := new(mockSpeccer)

		o := &Operator{
			ctx:     ctx,
			speccer: mockSpeccer,
		}

		currResource := &runtime.Unknown{}
		mockResource.On("Get", ctx, metav1.GetOptions{}).Once().Return(currResource, nil)

		proposedResource := &runtime.Unknown{}
		mockResource.On("Apply", ctx, metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}).
			Once().
			Return(proposedResource, nil)

		currSpec := "CURRENTSPEC"
		mockSpeccer.On("Spec", currResource).Once().Return(currSpec, nil)

		specErr := errors.New("couldnt spec second")
		mockSpeccer.On("Spec", currResource).Once().Return(nil, specErr)

		// Act
		neededModification, err := o.Apply(mockResource)

		// Assert
		assert.EqualError(t, err, specErr.Error())
		assert.False(t, neededModification)

		mockResource.AssertExpectations(t)
		mockSpeccer.AssertExpectations(t)
	})

	t.Run("Returns error if could not do actual apply", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockResource := new(mockResource)
		mockSpeccer := new(mockSpeccer)

		o := &Operator{
			ctx:     ctx,
			speccer: mockSpeccer,
		}

		currResource := &runtime.Unknown{}
		mockResource.On("Get", ctx, metav1.GetOptions{}).Once().Return(currResource, nil)

		proposedResource := &runtime.Unknown{}
		mockResource.On("Apply", ctx, metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}}).
			Once().
			Return(proposedResource, nil)

		currSpec := "CURRENTSPEC"
		mockSpeccer.On("Spec", currResource).Once().Return(currSpec, nil)

		proposedSpec := "PROPOSEDSPEC"
		mockSpeccer.On("Spec", currResource).Once().Return(proposedSpec, nil)

		applyErr := errors.New("couldnt do actual apply")
		mockResource.On("Apply", ctx, metav1.ApplyOptions{FieldManager: "goflow-cli"}).
			Once().
			Return(nil, applyErr)

		// Act
		neededModification, err := o.Apply(mockResource)

		// Assert
		assert.EqualError(t, err, applyErr.Error())
		assert.False(t, neededModification)

		mockResource.AssertExpectations(t)
		mockSpeccer.AssertExpectations(t)
	})
}

func Test_Operator_Delete(t *testing.T) {
	t.Run("Returns true if the resource was deleted successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockResource := new(mockResource)

		o := &Operator{
			ctx: ctx,
		}

		mockResource.On("Delete", ctx, metav1.DeleteOptions{}).Once().Return(nil)

		// Act
		neededDeletion, err := o.Delete(mockResource)

		// Assert
		assert.NoError(t, err)
		assert.True(t, neededDeletion)

		mockResource.AssertExpectations(t)
	})

	t.Run("Returns false if the resource does not exist", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockResource := new(mockResource)

		o := &Operator{
			ctx: ctx,
		}

		notFoundErr := k8serr.NewNotFound(schema.GroupResource{}, "test-resource")
		mockResource.On("Delete", ctx, metav1.DeleteOptions{}).Once().Return(notFoundErr)

		// Act
		neededDeletion, err := o.Delete(mockResource)

		// Assert
		assert.NoError(t, err)
		assert.False(t, neededDeletion)

		mockResource.AssertExpectations(t)
	})

	t.Run("Returns error if there was an error deleting the resource", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockResource := new(mockResource)

		o := &Operator{
			ctx: ctx,
		}

		deleteErr := errors.New("could not delete")
		mockResource.On("Delete", ctx, metav1.DeleteOptions{}).Once().Return(deleteErr)

		// Act
		neededDeletion, err := o.Delete(mockResource)

		// Assert
		assert.EqualError(t, err, deleteErr.Error())
		assert.False(t, neededDeletion)

		mockResource.AssertExpectations(t)
	})
}

func Test_Operator_WaitFor(t *testing.T) {
	t.Run("Doesn't return error if correct event type was found", func(t *testing.T) {
		// Arrange
		listOptions := metav1.ListOptions{}
		mockResource := new(mockResource)
		ctx := context.Background()
		o := Operator{
			ctx: ctx,
		}

		watcher := watch.NewFakeWithChanSize(1, false)
		mockResource.On("Watch", ctx, listOptions).Once().Return(watcher, nil)

		// Act
		watcher.Add(&runtime.Unknown{})

		err := o.WaitFor(mockResource, []watch.EventType{watch.Added}, time.Second)

		// Assert
		assert.NoError(t, err)
		mockResource.AssertExpectations(t)
	})

	t.Run("Returns error if couldn't watch resource", func(t *testing.T) {
		// Arrange
		listOptions := metav1.ListOptions{}
		mockResource := new(mockResource)
		ctx := context.Background()
		o := Operator{
			ctx: ctx,
		}

		watchErr := errors.New("couldnt watch")
		mockResource.On("Watch", ctx, listOptions).Once().Return(nil, watchErr)

		// Act
		err := o.WaitFor(mockResource, []watch.EventType{watch.Added}, time.Second)

		// Assert
		assert.EqualError(t, err, watchErr.Error())
		mockResource.AssertExpectations(t)
	})

	t.Run("Returns an error if timed out", func(t *testing.T) {
		// Arrange
		listOptions := metav1.ListOptions{}
		mockResource := new(mockResource)
		ctx := context.Background()
		o := Operator{
			ctx: ctx,
		}

		watcher := watch.NewFake()
		mockResource.On("Watch", ctx, listOptions).Once().Return(watcher, nil)

		// Act
		err := o.WaitFor(mockResource, []watch.EventType{watch.Added}, time.Millisecond)

		// Assert
		assert.EqualError(t, err, "timeout reached waiting for events")
		mockResource.AssertExpectations(t)
	})
	t.Run("Returns an error if watcher channel closed", func(t *testing.T) {
		// Arrange
		listOptions := metav1.ListOptions{}
		mockResource := new(mockResource)
		ctx := context.Background()
		o := Operator{
			ctx: ctx,
		}

		watcher := watch.NewFake()
		mockResource.On("Watch", ctx, listOptions).Once().Return(watcher, nil)

		// Act
		watcher.Stop()

		err := o.WaitFor(mockResource, []watch.EventType{watch.Added}, 10*time.Second)

		// Assert
		assert.EqualError(t, err, "watcher channel closed unexpectedly")
		mockResource.AssertExpectations(t)
	})
}

type mockResource struct {
	mock.Mock
}

func (m *mockResource) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockResource) Kind() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockResource) Apply(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
	args := m.Called(ctx, opts)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(runtime.Object), args.Error(1)
}

func (m *mockResource) Delete(ctx context.Context, opts metav1.DeleteOptions) error {
	args := m.Called(ctx, opts)

	return args.Error(0)
}

func (m *mockResource) Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
	args := m.Called(ctx, opts)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(runtime.Object), args.Error(1)
}

func (m *mockResource) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	args := m.Called(ctx, opts)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(watch.Interface), args.Error(1)
}

type mockSpeccer struct {
	mock.Mock
}

func (m *mockSpeccer) Spec(obj runtime.Object) (any, error) {
	args := m.Called(obj)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0), args.Error(1)
}

type mockKubeConfigBuilder struct {
	mock.Mock
}

func (m *mockKubeConfigBuilder) GetKubeConfigPath() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *mockKubeConfigBuilder) BuildConfig(clusterURL, kubeConfigPath string) (*rest.Config, error) {
	args := m.Called(clusterURL, kubeConfigPath)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*rest.Config), args.Error(1)
}

type mockClientsetBuilder struct {
	mock.Mock
}

func (m *mockClientsetBuilder) NewForConfig(config *rest.Config) (*kubernetes.Clientset, error) {
	args := m.Called(config)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*kubernetes.Clientset), args.Error(1)
}
