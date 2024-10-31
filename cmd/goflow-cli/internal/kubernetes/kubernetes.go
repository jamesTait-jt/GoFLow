package kubernetes

import (
	"context"
	"reflect"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/resource"
	"github.com/jamesTait-jt/goflow/pkg/log"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Resource interface {
	Name() string
	Kind() string

	Apply(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error)
	Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error)
}

type speccer interface {
	Spec(obj runtime.Object) (any, error)
}

type Operator struct {
	ctx    context.Context
	logger log.Logger
	// waiter            resource.EventWaiter
	speccer speccer
}

func NewOperator(opts ...OperatorOption) (*Operator, error) {
	options := defaultOperatorOptions()

	for _, o := range opts {
		o.apply(&options)
	}

	ctx := context.Background()

	// eventWaiter := resource.NewWaiter(ctx, options.logger)

	client := &Operator{
		ctx:    ctx,
		logger: options.logger,
		// waiter:           eventWaiter,
		speccer: &resource.ObjectSpeccer{},
	}

	return client, nil
}

func (k *Operator) Apply(r Resource) (bool, error) {
	currResource, err := r.Get(
		k.ctx,
		metav1.GetOptions{},
	)

	if err != nil && !k8serr.IsNotFound(err) {
		return false, err
	}

	proposedResource, err := r.Apply(
		k.ctx,
		metav1.ApplyOptions{FieldManager: "goflow-cli", DryRun: []string{"All"}},
	)

	if err != nil {
		return false, err
	}

	currSpec, err := k.speccer.Spec(currResource)
	if err != nil {
		return false, err
	}

	proposedSpec, err := k.speccer.Spec(proposedResource)
	if err != nil {
		return false, err
	}

	// new spec is the same as the old spec - no changes
	if reflect.DeepEqual(currSpec, proposedSpec) {
		return false, nil
	}

	_, err = r.Apply(
		k.ctx,
		metav1.ApplyOptions{FieldManager: "goflow-cli"},
	)

	if err != nil {
		return false, err
	}

	return true, nil
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
