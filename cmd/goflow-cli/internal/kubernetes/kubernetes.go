package kubernetes

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/resource"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/pkg/slice"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Applier interface {
	Apply(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error)
}

type Getter interface {
	Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error)
}

type Deleter interface {
	Delete(ctx context.Context, opts metav1.DeleteOptions) error
}

type Watcher interface {
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type ApplyGetter interface {
	Applier
	Getter
}

type speccer interface {
	Spec(obj runtime.Object) (any, error)
}

type Operator struct {
	ctx     context.Context
	logger  log.Logger
	speccer speccer
}

func NewOperator(opts ...OperatorOption) (*Operator, error) {
	options := defaultOperatorOptions()

	for _, o := range opts {
		o.apply(&options)
	}

	ctx := context.Background()

	client := &Operator{
		ctx:     ctx,
		logger:  options.logger,
		speccer: &resource.ObjectSpeccer{},
	}

	return client, nil
}

func (o *Operator) Apply(kubeResource ApplyGetter) (bool, error) {
	currResource, err := kubeResource.Get(o.ctx, metav1.GetOptions{})

	if err != nil && !k8serr.IsNotFound(err) {
		return false, err
	}

	proposedResource, err := kubeResource.Apply(
		o.ctx,
		metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}},
	)

	if err != nil {
		return false, err
	}

	currSpec, err := o.speccer.Spec(currResource)
	if err != nil {
		return false, err
	}

	proposedSpec, err := o.speccer.Spec(proposedResource)
	if err != nil {
		return false, err
	}

	// new spec is the same as the old spec - no changes
	if reflect.DeepEqual(currSpec, proposedSpec) {
		return false, nil
	}

	_, err = kubeResource.Apply(o.ctx, metav1.ApplyOptions{FieldManager: "goflow-cli"})

	return err == nil, err
}

func (o *Operator) Delete(kubeResource Deleter) (bool, error) {
	err := kubeResource.Delete(o.ctx, metav1.DeleteOptions{})

	if err != nil {
		// was not found - no need to delete
		if k8serr.IsNotFound(err) {
			return false, nil
		}

		// some other error occurred
		return false, err
	}

	// needed to delete
	return true, nil
}

func (o *Operator) WaitFor(kubeResource Watcher, eventTypes []watch.EventType, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(o.ctx, timeout)
	defer cancel()

	watcher, err := kubeResource.Watch(o.ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			// Timeout or cancellation happened
			return errors.New("timeout reached waiting for events")

		case event, ok := <-watcher.ResultChan():
			// Check if the channel was closed unexpectedly
			if !ok {
				return errors.New("watcher channel closed unexpectedly")
			}

			if slice.Contains(eventTypes, event.Type) {
				return nil
			}
		}
	}
}

type kubeConfigBuilder interface {
	GetKubeConfigPath() (string, error)
	BuildConfig(clusterURL, kubeConfigPath string) (*rest.Config, error)
}

type clientSetBuilder interface {
	NewForConfig(config *rest.Config) (*kubernetes.Clientset, error)
}

func NewClientset(clusterURL string, opts ...BuildClientsetOption) (kubernetes.Interface, error) {
	options := defaultBuildClientsetOptions()

	for _, o := range opts {
		o.apply(&options)
	}

	kubeConfigPath, err := options.configBuilder.GetKubeConfigPath()
	if err != nil {
		return nil, err
	}

	kubeConfig, err := options.configBuilder.BuildConfig(clusterURL, kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return options.kubeClientBuilder.NewForConfig(kubeConfig)
}
