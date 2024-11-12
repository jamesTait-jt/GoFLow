package k8s

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/k8s/grpcserver"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/k8s/redis"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/k8s/resource"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/k8s/workerpool"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

type ConfigMapper struct {
	conf *config.Config
}

func NewConfigMapper(conf *config.Config) *ConfigMapper {
	return &ConfigMapper{conf: conf}
}

func (cm *ConfigMapper) GetNamespaceConfig(key resource.Key) (*acapiv1.NamespaceApplyConfiguration, error) {
	switch key {
	case resource.Namespace:
		return acapiv1.Namespace(cm.conf.Kubernetes.Namespace), nil

	default:
		return nil, fmt.Errorf("didnt recognise namespace resource key '%d'", key)
	}
}

func (cm *ConfigMapper) GetDeploymentConfig(key resource.Key) (*acappsv1.DeploymentApplyConfiguration, error) {
	switch key {
	case resource.GRPCServerDeployment:
		return grpcserver.Deployment(cm.conf), nil

	case resource.MessageBrokerDeployment:
		return redis.Deployment(cm.conf), nil

	case resource.WorkerpoolDeployment:
		return workerpool.Deployment(cm.conf), nil

	default:
		return nil, fmt.Errorf("didnt recognise deployment resource key '%d'", key)
	}
}

func (cm *ConfigMapper) GetServiceConfig(key resource.Key) (*acapiv1.ServiceApplyConfiguration, error) {
	switch key {
	case resource.GRPCServerService:
		return grpcserver.Service(cm.conf), nil

	case resource.MessageBrokerService:
		return redis.Service(cm.conf), nil

	default:
		return nil, fmt.Errorf("didnt recognise service resource key '%d'", key)
	}
}

func (cm *ConfigMapper) GetPersistentVolumeConfig(key resource.Key) (*acapiv1.PersistentVolumeApplyConfiguration, error) {
	switch key {
	case resource.WorkerpoolPV:
		return workerpool.HandlersPV(cm.conf), nil

	default:
		return nil, fmt.Errorf("didnt recognise persistent volume resource key '%d'", key)
	}
}

func (cm *ConfigMapper) GetPersistentVolumeClaimConfig(key resource.Key) (*acapiv1.PersistentVolumeClaimApplyConfiguration, error) {
	switch key {
	case resource.WorkerpoolPVC:
		return workerpool.HandlersPVC(cm.conf), nil

	default:
		return nil, fmt.Errorf("didnt recognise persistent volume claim resource key '%d'", key)
	}
}
