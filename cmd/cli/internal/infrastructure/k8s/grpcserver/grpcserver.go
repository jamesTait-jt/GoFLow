package grpcserver

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/infrastructure/k8s/redis"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
	acmetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

var (
	deploymentName                = "goflow-grpc-deployment"
	deploymentContainerName       = "goflow-grpc-deployment-container"
	serviceName                   = "goflow-grpc-service"
	GRPCPort                int32 = 50051

	labels = map[string]string{
		"app": "goflow-grpc-server",
	}
)

func Deployment(conf *config.Config) *acappsv1.DeploymentApplyConfiguration {
	return acappsv1.Deployment(deploymentName, conf.Kubernetes.Namespace).
		WithLabels(labels).
		WithSpec(
			acappsv1.DeploymentSpec().
				WithReplicas(conf.GoFlowServer.Replicas).
				WithSelector(acmetav1.LabelSelector().WithMatchLabels(labels)).
				WithTemplate(
					acapiv1.PodTemplateSpec().
						WithLabels(labels).
						WithSpec(
							acapiv1.PodSpec().
								WithContainers(
									acapiv1.Container().
										WithName(deploymentContainerName).
										WithImage(conf.GoFlowServer.Image).
										WithImagePullPolicy(apiv1.PullIfNotPresent).
										WithArgs(
											"--broker-type", "redis",
											"--broker-addr", fmt.Sprintf("%s:%d", redis.ServiceName, redis.RedisPort),
										).
										WithPorts(
											acapiv1.ContainerPort().
												WithProtocol(
													apiv1.ProtocolTCP,
												).WithContainerPort(
												GRPCPort,
											),
										),
								),
						),
				),
		)
}

func Service(conf *config.Config) *acapiv1.ServiceApplyConfiguration {
	return acapiv1.Service(serviceName, conf.Kubernetes.Namespace).
		WithLabels(labels).
		WithSpec(
			acapiv1.ServiceSpec().
				WithSelector(labels).
				WithType(apiv1.ServiceTypeLoadBalancer).
				WithLoadBalancerIP(conf.GoFlowServer.Address).
				WithPorts(
					acapiv1.ServicePort().
						WithPort(GRPCPort).
						WithTargetPort(intstr.FromInt32(GRPCPort)),
				),
		)
}
