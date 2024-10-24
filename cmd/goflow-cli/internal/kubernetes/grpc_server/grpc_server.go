package grpcserver

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	deploymentName                = "goflow-grpc-deployment"
	deploymentContainerName       = "goflow-grpc-deployment-container"
	serviceName                   = "goflow-grpc-service"
	grpcPort                int32 = 50051

	labels = map[string]string{
		"app": "goflow-grpc-server",
	}
)

func Deployment(conf *config.Config) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: conf.Kubernetes.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &conf.GoFlowServer.Replicas,
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
							Name:            deploymentContainerName,
							Image:           conf.GoFlowServer.Image,
							ImagePullPolicy: apiv1.PullNever,
							Ports: []apiv1.ContainerPort{
								{
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: grpcPort,
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
			Name:      serviceName,
			Labels:    labels,
			Namespace: conf.Kubernetes.Namespace,
		},
		Spec: apiv1.ServiceSpec{
			Selector:       labels,
			Type:           apiv1.ServiceTypeLoadBalancer,
			LoadBalancerIP: conf.GoFlowServer.Address,
			Ports: []apiv1.ServicePort{
				{
					Port:       grpcPort,
					TargetPort: intstr.FromInt32(grpcPort),
				},
			},
		},
	}
}
