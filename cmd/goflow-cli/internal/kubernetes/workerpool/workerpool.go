package workerpool

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/kubernetes/redis"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	accorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	acmetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

var (
	pvName           = "handlers-pv"
	pvcName          = "handlers-pvc"
	storageClassName = "standard"

	volumeMountName            = "handlers-volume-mount"
	pluginBuilderContainerName = "plugin-builder-container"
	workerpoolContainerName    = "workerpool-container"
	deploymentName             = "workerpool-deployment"

	labels = map[string]string{
		"app": "goflow-workerpool",
	}
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

func Deployment(conf *config.Config) *acappsv1.DeploymentApplyConfiguration {
	volumeMount := accorev1.VolumeMount().WithName(
		volumeMountName,
	).WithMountPath(
		"/app/handlers",
	)

	pluginBuilderContainer := accorev1.Container().WithName(
		pluginBuilderContainerName,
	).WithImage(
		conf.Workerpool.PluginBuilderImage,
	).WithImagePullPolicy(
		apiv1.PullNever,
	).WithArgs(
		"/app/handlers",
	).WithVolumeMounts(
		volumeMount,
	)

	workerpoolContainer := accorev1.Container().WithName(
		workerpoolContainerName,
	).WithImage(
		conf.Workerpool.Image,
	).WithImagePullPolicy(
		apiv1.PullNever,
	).WithArgs(
		"--broker-type", "redis",
		"--broker-addr", fmt.Sprintf("%s:%d", redis.ServiceName, redis.RedisPort),
		"--handlers-path", "/app/handlers/compiled",
	).WithVolumeMounts(
		volumeMount,
	)

	return acappsv1.Deployment(
		deploymentName, conf.Kubernetes.Namespace,
	).WithLabels(
		labels,
	).WithSpec(
		acappsv1.DeploymentSpec().WithReplicas(
			conf.Workerpool.Replicas,
		).WithSelector(
			acmetav1.LabelSelector().WithMatchLabels(labels),
		).WithTemplate(
			accorev1.PodTemplateSpec().WithLabels(
				labels,
			).WithSpec(
				accorev1.PodSpec().WithRestartPolicy(
					apiv1.RestartPolicyAlways,
				).WithInitContainers(
					pluginBuilderContainer,
				).WithContainers(
					workerpoolContainer,
				).WithVolumes(
					accorev1.Volume().WithName(
						*volumeMount.Name,
					).WithPersistentVolumeClaim(
						accorev1.PersistentVolumeClaimVolumeSource().WithClaimName(pvcName),
					),
				),
			),
		),
	)
}
