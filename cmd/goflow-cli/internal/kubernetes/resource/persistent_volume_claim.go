package resource

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedapiv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type PersistentVolumeClaim struct {
	config *acapiv1.PersistentVolumeClaimApplyConfiguration
	client typedapiv1.PersistentVolumeClaimInterface
}

func NewPersistentVolumeClaim(config *acapiv1.PersistentVolumeClaimApplyConfiguration, client typedapiv1.PersistentVolumeClaimInterface) *PersistentVolumeClaim {
	return &PersistentVolumeClaim{
		config: config,
		client: client,
	}
}

func (pvc *PersistentVolumeClaim) Name() string {
	return *pvc.config.Name
}

func (pvc *PersistentVolumeClaim) Kind() string {
	return "pvc"
}

func (pvc *PersistentVolumeClaim) Apply(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
	return pvc.client.Apply(ctx, pvc.config, opts)
}

func (pvc *PersistentVolumeClaim) Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
	return pvc.client.Get(ctx, pvc.Name(), opts)
}
