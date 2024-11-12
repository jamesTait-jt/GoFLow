//go:build unit

package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

func Test_ObjectSpeccer_Spec(t *testing.T) {
	t.Run("Returns spec for namespace", func(t *testing.T) {
		// Arrange
		s := &ObjectSpeccer{}

		spec := apiv1.NamespaceSpec{}
		namespace := &apiv1.Namespace{
			Spec: spec,
		}

		// Act
		returnedSpec, err := s.Spec(namespace)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, spec, returnedSpec)
	})

	t.Run("Returns spec for deployment", func(t *testing.T) {
		// Arrange
		s := &ObjectSpeccer{}

		spec := appsv1.DeploymentSpec{}
		deployment := &appsv1.Deployment{Spec: spec}

		// Act
		returnedSpec, err := s.Spec(deployment)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, spec, returnedSpec)
	})

	t.Run("Returns spec for service", func(t *testing.T) {
		// Arrange
		s := &ObjectSpeccer{}

		spec := apiv1.ServiceSpec{}
		service := &apiv1.Service{Spec: spec}

		// Act
		returnedSpec, err := s.Spec(service)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, spec, returnedSpec)
	})

	t.Run("Returns spec for persistent volume", func(t *testing.T) {
		// Arrange
		s := &ObjectSpeccer{}

		spec := apiv1.PersistentVolumeSpec{}
		pv := &apiv1.PersistentVolume{Spec: spec}

		// Act
		returnedSpec, err := s.Spec(pv)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, spec, returnedSpec)
	})

	t.Run("Returns spec for persistent volume claim", func(t *testing.T) {
		// Arrange
		s := &ObjectSpeccer{}

		spec := apiv1.PersistentVolumeClaimSpec{}
		pvc := &apiv1.PersistentVolumeClaim{Spec: spec}

		// Act
		returnedSpec, err := s.Spec(pvc)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, spec, returnedSpec)
	})

	t.Run("Returns error for unsupported object type", func(t *testing.T) {
		// Arrange
		s := &ObjectSpeccer{}

		unsupportedObj := &apiv1.Pod{} // Using Pod as an example of an unsupported type

		// Act
		returnedSpec, err := s.Spec(unsupportedObj)

		// Assert
		assert.NotNil(t, err)
		assert.Nil(t, returnedSpec)
		assert.Contains(t, err.Error(), "couldn't get spec of unrecognised object")
	})
}
