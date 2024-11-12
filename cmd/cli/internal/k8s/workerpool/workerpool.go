package workerpool

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/k8s/redis"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
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

func HandlersPV(conf *config.Config) *acapiv1.PersistentVolumeApplyConfiguration {
	return acapiv1.PersistentVolume(pvName).
		WithSpec(
			acapiv1.PersistentVolumeSpec().
				WithCapacity(apiv1.ResourceList{apiv1.ResourceStorage: resource.MustParse("1Gi")}).
				WithAccessModes(apiv1.ReadWriteMany).
				WithStorageClassName(storageClassName).
				WithHostPath(
					acapiv1.HostPathVolumeSource().
						WithPath(conf.Workerpool.PathToHandlers),
				),
		)
}

func HandlersPVC(conf *config.Config) *acapiv1.PersistentVolumeClaimApplyConfiguration {
	return acapiv1.PersistentVolumeClaim(pvcName, conf.Kubernetes.Namespace).
		WithSpec(
			acapiv1.PersistentVolumeClaimSpec().
				WithVolumeName(pvName).
				WithAccessModes(apiv1.ReadWriteMany).
				WithStorageClassName(storageClassName).
				WithResources(
					acapiv1.VolumeResourceRequirements().
						WithRequests(apiv1.ResourceList{apiv1.ResourceStorage: resource.MustParse("1Gi")}),
				),
		)
}

func Deployment(conf *config.Config) *acappsv1.DeploymentApplyConfiguration {
	volumeMount := acapiv1.VolumeMount().
		WithName(volumeMountName).
		WithMountPath("/app/handlers")

	pluginBuilderContainer := acapiv1.Container().
		WithName(pluginBuilderContainerName).
		WithImage(conf.Workerpool.PluginBuilderImage).
		WithImagePullPolicy(apiv1.PullNever).
		WithArgs("/app/handlers").
		WithVolumeMounts(volumeMount)

	workerpoolContainer := acapiv1.Container().
		WithName(workerpoolContainerName).
		WithImage(conf.Workerpool.Image).
		WithImagePullPolicy(apiv1.PullNever).
		WithVolumeMounts(volumeMount).
		WithArgs(
			"--broker-type", "redis",
			"--broker-addr", fmt.Sprintf("%s:%d", redis.ServiceName, redis.RedisPort),
			"--handlers-path", "/app/handlers/compiled",
		)

	return acappsv1.Deployment(deploymentName, conf.Kubernetes.Namespace).
		WithLabels(labels).
		WithSpec(
			acappsv1.DeploymentSpec().
				WithReplicas(conf.Workerpool.Replicas).
				WithSelector(acmetav1.LabelSelector().WithMatchLabels(labels)).
				WithTemplate(
					acapiv1.PodTemplateSpec().
						WithLabels(labels).
						WithSpec(
							acapiv1.PodSpec().
								WithRestartPolicy(apiv1.RestartPolicyAlways).
								WithInitContainers(pluginBuilderContainer).
								WithContainers(workerpoolContainer).
								WithVolumes(
									acapiv1.Volume().
										WithName(*volumeMount.Name).
										WithPersistentVolumeClaim(
											acapiv1.PersistentVolumeClaimVolumeSource().
												WithClaimName(pvcName),
										),
								),
						),
				),
		)
}
