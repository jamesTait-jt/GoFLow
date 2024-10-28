package grpcserver

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
	deploymentName                = "goflow-grpc-deployment"
	deploymentContainerName       = "goflow-grpc-deployment-container"
	serviceName                   = "goflow-grpc-service"
	grpcPort                int32 = 50051

	labels = map[string]string{
		"app": "goflow-grpc-server",
	}
)

func Deployment(conf *config.Config) *acappsv1.DeploymentApplyConfiguration {
	return acappsv1.Deployment(
		deploymentName, conf.Kubernetes.Namespace,
	).WithLabels(
		labels,
	).WithSpec(
		acappsv1.DeploymentSpec().WithReplicas(
			conf.GoFlowServer.Replicas,
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
						conf.GoFlowServer.Image,
					).WithImagePullPolicy(
						apiv1.PullNever,
					).WithPorts(
						accorev1.ContainerPort().WithProtocol(
							apiv1.ProtocolTCP,
						).WithContainerPort(
							grpcPort,
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
