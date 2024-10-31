package resource

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedapiv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Service struct {
	config *acapiv1.ServiceApplyConfiguration
	client typedapiv1.ServiceInterface
}

func NewService(config *acapiv1.ServiceApplyConfiguration, client typedapiv1.ServiceInterface) *Service {
	return &Service{
		config: config,
		client: client,
	}
}

func (s *Service) Name() string {
	return *s.config.Name
}

func (s *Service) Kind() string {
	return "service"
}

func (s *Service) Apply(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
	return s.client.Apply(ctx, s.config, opts)
}

func (s *Service) Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
	return s.client.Get(ctx, s.Name(), opts)
}
