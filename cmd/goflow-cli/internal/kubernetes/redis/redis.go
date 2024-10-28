package redis

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	accorev1 "k8s.io/client-go/applyconfigurations/core/v1"
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
	return acappsv1.Deployment(
		deploymentName, conf.Kubernetes.Namespace,
	).WithLabels(
		labels,
	).WithSpec(
		acappsv1.DeploymentSpec().WithReplicas(
			conf.Redis.Replicas,
		).WithSelector(
			acmetav1.LabelSelector().WithMatchLabels(labels),
		).WithTemplate(
			accorev1.PodTemplateSpec().WithLabels(
				labels,
			).WithSpec(
				accorev1.PodSpec().WithContainers(
					accorev1.Container().WithName(
						deploymentContainerName,
					).WithImage(
						conf.Redis.Image,
					).WithPorts(
						accorev1.ContainerPort().WithProtocol(
							apiv1.ProtocolTCP,
						).WithContainerPort(
							RedisPort,
						),
					),
				),
			),
		),
	)
}

func Service(conf *config.Config) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Labels:    labels,
			Namespace: conf.Kubernetes.Namespace,
		},
		Spec: apiv1.ServiceSpec{
			Selector: labels,
			Type:     apiv1.ServiceTypeClusterIP,
			Ports: []apiv1.ServicePort{
				{
					Port:       RedisPort,
					TargetPort: intstr.FromInt32(RedisPort),
				},
			},
		},
	}
}
