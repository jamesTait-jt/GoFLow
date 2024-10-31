package kubernetes

import (
	"context"
	"errors"
	"testing"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/resource"
	"github.com/jamesTait-jt/goflow/pkg/log"
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
		mockLogger := new(log.TestifyMock)

		// Act
		kube, err := NewOperator(
			WithLogger(mockLogger),
		)

		// Assert
		assert.Nil(t, err)
		assert.NotNil(t, kube.ctx)
		assert.Equal(t, mockLogger, kube.logger)
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

func Test_NewClientSet(t *testing.T) {
	t.Run("Returns clientset", func(t *testing.T) {
		// Arrange
		clusterURL := "cluster"
		mockKubeConfBuilder := new(mockKubeConfigBuilder)
		mockClientsetBuilder := new(mockClientsetBuilder)

		kubeconfigPath := "config path"
		mockKubeConfBuilder.On("GetKubeConfigPath").Once().Return(kubeconfigPath, nil)

		kubeConfig := &rest.Config{}
		mockKubeConfBuilder.On("BuildConfig", clusterURL, kubeconfigPath).Once().Return(kubeConfig, nil)

		clientset := &kubernetes.Clientset{}
		mockClientsetBuilder.On("NewForConfig", kubeConfig).Once().Return(clientset, nil)

		// Act
		cs, err := NewClientset(
			clusterURL,
			WithConfigBuilder(mockKubeConfBuilder),
			WithKubeClientBuilder(mockClientsetBuilder),
		)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, clientset, cs)

		mockKubeConfBuilder.AssertExpectations(t)
		mockClientsetBuilder.AssertExpectations(t)
	})

	t.Run("Returns error if couldnt get kube conf path", func(t *testing.T) {
		// Arrange
		clusterURL := "cluster"
		mockKubeConfBuilder := new(mockKubeConfigBuilder)
		mockClientsetBuilder := new(mockClientsetBuilder)

		getPathErr := errors.New("couldnt get path")
		mockKubeConfBuilder.On("GetKubeConfigPath").Once().Return("", getPathErr)

		// Act
		cs, err := NewClientset(
			clusterURL,
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
		mockKubeConfBuilder := new(mockKubeConfigBuilder)
		mockClientsetBuilder := new(mockClientsetBuilder)

		kubeconfigPath := "config path"
		mockKubeConfBuilder.On("GetKubeConfigPath").Once().Return(kubeconfigPath, nil)

		buildConfErr := errors.New("couldnt build conf")
		mockKubeConfBuilder.On("BuildConfig", clusterURL, kubeconfigPath).Once().Return(nil, buildConfErr)

		// Act
		cs, err := NewClientset(
			clusterURL,
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

func (m *mockResource) Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
	args := m.Called(ctx, opts)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(runtime.Object), args.Error(1)
}

type mockSpeccer struct {
	mock.Mock
}

func (m *mockSpeccer) Spec(obj runtime.Object) (any, error) {
	args := m.Called(obj)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(any), args.Error(1)
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

type MockEventWaiter struct {
	mock.Mock
}

func (m *MockEventWaiter) WaitFor(resourceName, namespace string, eventTypes []watch.EventType, client resource.Watchable) error {
	args := m.Called(resourceName, namespace, eventTypes, client)
	return args.Error(0)
}
