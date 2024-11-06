package resource

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/grpcserver"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/redis"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/workerpool"

	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type Clientset interface {
	Namespaces() NamespaceInterface
	Deployments() DeploymentInterface
	Services() ServiceInterface
	PersistentVolumes() PersistentVolumeInterface
	PersistentVolumeClaims() PersistentVolumeClaimInterface
}

type messageBrokerApplyConfigs struct {
	deployment *acappsv1.DeploymentApplyConfiguration
	service    *acapiv1.ServiceApplyConfiguration
}

type gRPCServerApplyConfigs struct {
	deployment *acappsv1.DeploymentApplyConfiguration
	service    *acapiv1.ServiceApplyConfiguration
}

type workerpoolApplyConfigs struct {
	pv         *acapiv1.PersistentVolumeApplyConfiguration
	pvc        *acapiv1.PersistentVolumeClaimApplyConfiguration
	deployment *acappsv1.DeploymentApplyConfiguration
}

type Builder struct {
	clients                   Clientset
	namespaceApplyConfig      *acapiv1.NamespaceApplyConfiguration
	messageBrokerApplyConfigs messageBrokerApplyConfigs
	gRPCServerApplyConfigs    gRPCServerApplyConfigs
	workerpoolApplyConfigs    workerpoolApplyConfigs
}

// TODO: Interface for getting the apply configs?
func NewBuilder(conf *config.Config, clients Clientset) *Builder {
	return &Builder{
		clients:              clients,
		namespaceApplyConfig: acapiv1.Namespace(conf.Kubernetes.Namespace),
		messageBrokerApplyConfigs: messageBrokerApplyConfigs{
			deployment: redis.Deployment(conf),
			service:    redis.Service(conf),
		},
		gRPCServerApplyConfigs: gRPCServerApplyConfigs{
			deployment: grpcserver.Deployment(conf),
			service:    grpcserver.Service(conf),
		},
		workerpoolApplyConfigs: workerpoolApplyConfigs{
			pv:         workerpool.HandlersPV(conf),
			pvc:        workerpool.HandlersPVC(conf),
			deployment: workerpool.Deployment(conf),
		},
	}
}

func (b *Builder) Build(key Key) *Resource {
	switch key {
	case Namespace:
		return NewNamespace(b.namespaceApplyConfig, b.clients.Namespaces())

	case MessageBrokerDeployment:
		return NewDeployment(b.messageBrokerApplyConfigs.deployment, b.clients.Deployments())

	case MessageBrokerService:
		return NewService(b.messageBrokerApplyConfigs.service, b.clients.Services())

	case GRPCServerDeployment:
		return NewDeployment(b.gRPCServerApplyConfigs.deployment, b.clients.Deployments())

	case GRPCServerService:
		return NewService(b.gRPCServerApplyConfigs.service, b.clients.Services())

	case WorkerpoolDeployment:
		return NewDeployment(b.workerpoolApplyConfigs.deployment, b.clients.Deployments())

	case WorkerpoolPV:
		return NewPersistentVolume(b.workerpoolApplyConfigs.pv, b.clients.PersistentVolumes())

	case WorkerpoolPVC:
		return NewPersistentVolumeClaim(b.workerpoolApplyConfigs.pvc, b.clients.PersistentVolumeClaims())
	}

	return nil
}
