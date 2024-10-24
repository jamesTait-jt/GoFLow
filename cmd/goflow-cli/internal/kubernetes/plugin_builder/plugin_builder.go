package pluginbuilder

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	volumeName      = "plugin-builder-volume"
	volumeMountName = "plugin-builder-volume-mount"
	mountedLocation = "/app/handlers"

	containerName = "plugin-builder-container"

	jobName = "plugin-builder-job"
)

func Job(conf *config.Config) *batchv1.Job {
	volume := apiv1.Volume{
		Name: volumeName,
		VolumeSource: apiv1.VolumeSource{
			HostPath: &apiv1.HostPathVolumeSource{
				Path: conf.Workerpool.PathToHandlers,
			},
		},
	}

	volumeMount := apiv1.VolumeMount{
		Name:      volumeMountName,
		MountPath: mountedLocation,
	}

	container := apiv1.Container{
		Name:  containerName,
		Image: conf.Workerpool.PluginBuilderImage,
		Args:  []string{mountedLocation},
		VolumeMounts: []apiv1.VolumeMount{
			volumeMount,
		},
	}

	four := int32(4)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers:    []apiv1.Container{container},
					Volumes:       []apiv1.Volume{volume},
					RestartPolicy: apiv1.RestartPolicyOnFailure,
				},
			},
			BackoffLimit: &four, // Retry up to 4 times if Job fails
		},
	}
}
