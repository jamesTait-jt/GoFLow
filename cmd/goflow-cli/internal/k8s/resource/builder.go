package resource

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/grpcserver"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/redis"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/workerpool"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type Clientset interface {
	Namespaces() NamespaceInterface
	Deployments() DeploymentInterface
	Services() ServiceInterface
	PersistentVolumes() PersistentVolumeInterface
	PersistentVolumeClaims() PersistentVolumeClaimInterface
}

type Builder struct {
	conf    *config.Config
	clients Clientset
}

func NewBuilder(conf *config.Config, clients Clientset) *Builder {
	return &Builder{
		conf:    conf,
		clients: clients,
	}
}

func (b *Builder) Build(key Key) *Resource {
	switch key {
	case Namespace:
		return NewNamespace(acapiv1.Namespace(b.conf.Kubernetes.Namespace), b.clients.Namespaces())

	case MessageBrokerDeployment:
		return NewDeployment(redis.Deployment(b.conf), b.clients.Deployments())

	case MessageBrokerService:
		return NewService(redis.Service(b.conf), b.clients.Services())

	case GRPCServerDeployment:
		return NewDeployment(grpcserver.Deployment(b.conf), b.clients.Deployments())

	case GRPCServerService:
		return NewService(grpcserver.Service(b.conf), b.clients.Services())

	case WorkerpoolDeployment:
		return NewDeployment(workerpool.Deployment(b.conf), b.clients.Deployments())

	case WorkerpoolPV:
		return NewPersistentVolume(workerpool.HandlersPV(b.conf), b.clients.PersistentVolumes())

	case WorkerpoolPVC:
		return NewPersistentVolumeClaim(workerpool.HandlersPVC(b.conf), b.clients.PersistentVolumeClaims())
	}

	return nil
}
