package redis

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func Deployment(conf *config.Config) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: conf.Kubernetes.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &conf.Redis.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  deploymentContainerName,
							Image: conf.Redis.Image,
							Ports: []apiv1.ContainerPort{
								{
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: RedisPort,
								},
							},
						},
					},
				},
			},
		},
	}
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
