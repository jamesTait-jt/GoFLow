package resource

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ObjectSpeccer struct{}

func (o *ObjectSpeccer) Spec(obj runtime.Object) (any, error) {
	switch typedObj := obj.(type) {
	case *apiv1.Namespace:
		return typedObj.Spec, nil
	case *appsv1.Deployment:
		return typedObj.Spec, nil
	case *apiv1.Service:
		return typedObj.Spec, nil
	case *apiv1.PersistentVolume:
		return typedObj.Spec, nil
	case *apiv1.PersistentVolumeClaim:
		return typedObj.Spec, nil
	}

	return nil, fmt.Errorf("couldn't get spec of unrecognised object: %v", obj)
}
