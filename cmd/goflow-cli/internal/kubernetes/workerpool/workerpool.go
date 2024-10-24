package workerpool

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/redis"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	pvName           = "handlers-pv"
	pvcName          = "handlers-pvc"
	storageClassName = "standard"

	volumeMountName            = "handlers-volume-mount"
	pluginBuilderContainerName = "plugin-builder-container"
	workerpoolContainerName    = "workerpool-container"
	jobName                    = "workerpool-job"
)

func HandlersPV(conf *config.Config) *apiv1.PersistentVolume {
	return &apiv1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvName,
		},
		Spec: apiv1.PersistentVolumeSpec{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceStorage: resource.MustParse("1Gi"),
			},
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteMany,
			},
			PersistentVolumeSource: apiv1.PersistentVolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: conf.Workerpool.PathToHandlers,
				},
			},
			StorageClassName: storageClassName,
		},
	}
}

func HandlersPVC(conf *config.Config) *apiv1.PersistentVolumeClaim {
	return &apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: conf.Kubernetes.Namespace,
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			VolumeName: pvName,
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteMany,
			},
			Resources: apiv1.VolumeResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
			StorageClassName: &storageClassName,
		},
	}
}

func Job(conf *config.Config) *batchv1.Job {
	pluginBuilderContainer := apiv1.Container{
		Name:            pluginBuilderContainerName,
		Image:           conf.Workerpool.PluginBuilderImage,
		ImagePullPolicy: apiv1.PullNever,
		Args:            []string{"/app/handlers"},
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:      volumeMountName,
				MountPath: "/app/handlers",
			},
		},
	}

	workerpoolContainer := apiv1.Container{
		Name:            workerpoolContainerName,
		Image:           conf.Workerpool.Image,
		ImagePullPolicy: apiv1.PullNever,
		Args: []string{
			"--broker-type", "redis",
			"--broker-addr", fmt.Sprintf("%s:%d", redis.ServiceName, redis.RedisPort),
			"--handlers-path", "/app/handlers/compiled",
		},
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:      volumeMountName,
				MountPath: "/app/handlers",
			},
		},
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					RestartPolicy: apiv1.RestartPolicyNever,
					InitContainers: []apiv1.Container{
						pluginBuilderContainer,
					},
					Containers: []apiv1.Container{
						workerpoolContainer,
					},
					Volumes: []apiv1.Volume{
						{
							Name: volumeMountName,
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
			},
		},
	}
}
