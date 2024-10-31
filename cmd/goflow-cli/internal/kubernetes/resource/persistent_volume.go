package resource

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedapiv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type PersistentVolume struct {
	config *acapiv1.PersistentVolumeApplyConfiguration
	client typedapiv1.PersistentVolumeInterface
}

func NewPersistentVolume(config *acapiv1.PersistentVolumeApplyConfiguration, client typedapiv1.PersistentVolumeInterface) *PersistentVolume {
	return &PersistentVolume{
		config: config,
		client: client,
	}
}

func (pv *PersistentVolume) Name() string {
	return *pv.config.Name
}

func (pv *PersistentVolume) Kind() string {
	return "pv"
}

func (pv *PersistentVolume) Apply(ctx context.Context, opts metav1.ApplyOptions) (runtime.Object, error) {
	return pv.client.Apply(ctx, pv.config, opts)
}

func (pv *PersistentVolume) Get(ctx context.Context, opts metav1.GetOptions) (runtime.Object, error) {
	return pv.client.Get(ctx, pv.Name(), opts)
}
