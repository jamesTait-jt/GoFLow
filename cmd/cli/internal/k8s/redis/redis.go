package redis

import (
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
	acmetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

var (
	deploymentName                = "goflow-redis-deployment"
	deploymentContainerName       = "goflow-redis-deployment-container"
	ServiceName                   = "goflow-redis-server"
	RedisPort               int32 = 6379

	labels = map[string]string{
		"app": "goflow-redis",
	}
)

func Deployment(conf *config.Config) *acappsv1.DeploymentApplyConfiguration {
	return acappsv1.Deployment(deploymentName, conf.Kubernetes.Namespace).
		WithLabels(labels).
		WithSpec(
			acappsv1.DeploymentSpec().
				WithReplicas(conf.Redis.Replicas).
				WithSelector(acmetav1.LabelSelector().WithMatchLabels(labels)).
				WithTemplate(
					acapiv1.PodTemplateSpec().
						WithLabels(labels).
						WithSpec(
							acapiv1.PodSpec().
								WithContainers(
									acapiv1.Container().
										WithName(deploymentContainerName).
										WithImage(conf.Redis.Image).
										WithImagePullPolicy(apiv1.PullIfNotPresent).
										WithPorts(
											acapiv1.ContainerPort().
												WithProtocol(apiv1.ProtocolTCP).
												WithContainerPort(RedisPort),
										),
								),
						),
				),
		)
}

func Service(conf *config.Config) *acapiv1.ServiceApplyConfiguration {
	return acapiv1.Service(ServiceName, conf.Kubernetes.Namespace).
		WithLabels(labels).
		WithSpec(
			acapiv1.ServiceSpec().
				WithSelector(labels).
				WithType(apiv1.ServiceTypeClusterIP).
				WithPorts(
					acapiv1.ServicePort().
						WithPort(RedisPort).
						WithTargetPort(intstr.FromInt32(RedisPort)),
				),
		)
}
