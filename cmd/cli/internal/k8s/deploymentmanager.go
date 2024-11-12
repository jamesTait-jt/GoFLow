package k8s

import (
	"time"

	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/k8s/resource"
	"github.com/jamesTait-jt/goflow/pkg/log"
	acappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"

	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

var timeout = 30 * time.Second

type configMapper interface {
	GetNamespaceConfig(resourceKey resource.Key) (*acapiv1.NamespaceApplyConfiguration, error)
	GetDeploymentConfig(resourceKey resource.Key) (*acappsv1.DeploymentApplyConfiguration, error)
	GetServiceConfig(resourceKey resource.Key) (*acapiv1.ServiceApplyConfiguration, error)
	GetPersistentVolumeConfig(resourceKey resource.Key) (*acapiv1.PersistentVolumeApplyConfiguration, error)
	GetPersistentVolumeClaimConfig(resourceKey resource.Key) (*acapiv1.PersistentVolumeClaimApplyConfiguration, error)
}

type resourceFactory interface {
	CreateNamespace(config *acapiv1.NamespaceApplyConfiguration) *resource.Resource
	CreateDeployment(config *acappsv1.DeploymentApplyConfiguration) *resource.Resource
	CreateService(config *acapiv1.ServiceApplyConfiguration) *resource.Resource
	CreatePersistentVolume(config *acapiv1.PersistentVolumeApplyConfiguration) *resource.Resource
	CreatePersistentVolumeClaim(config *acapiv1.PersistentVolumeClaimApplyConfiguration) *resource.Resource
}

type deploymentExecutor interface {
	ApplyAndWait(kubeResource identifiableWatchableApplyGetter, timeout time.Duration) error
	DestroyAndWait(kubeResource identifiableWatchableDeleter, timeout time.Duration) error
}

type DeploymentManager struct {
	logger          log.Logger
	configMapper    configMapper
	resourceFactory resourceFactory
	executor        deploymentExecutor
}

func NewDeploymentManager(conf *config.Config, logger log.Logger, clients resource.Clientset) *DeploymentManager {
	return &DeploymentManager{
		logger:          logger,
		configMapper:    NewConfigMapper(conf),
		resourceFactory: resource.NewFactory(clients),
		executor:        NewDeploymentExecutor(logger),
	}
}

func (d *DeploymentManager) DeployNamespace() error {
	namespaceConf, err := d.configMapper.GetNamespaceConfig(resource.Namespace)
	if err != nil {
		return err
	}

	namespace := d.resourceFactory.CreateNamespace(namespaceConf)

	return d.executor.ApplyAndWait(namespace, timeout)
}

func (d *DeploymentManager) DeployMessageBroker() error {
	deploymentConf, err := d.configMapper.GetDeploymentConfig(resource.MessageBrokerDeployment)
	if err != nil {
		return err
	}

	deployment := d.resourceFactory.CreateDeployment(deploymentConf)

	if err = d.executor.ApplyAndWait(deployment, timeout); err != nil {
		return err
	}

	serviceConf, err := d.configMapper.GetServiceConfig(resource.MessageBrokerService)
	if err != nil {
		return err
	}

	service := d.resourceFactory.CreateService(serviceConf)

	return d.executor.ApplyAndWait(service, timeout)
}

func (d *DeploymentManager) DeployGRPCServer() error {
	deploymentConf, err := d.configMapper.GetDeploymentConfig(resource.GRPCServerDeployment)
	if err != nil {
		return err
	}

	deployment := d.resourceFactory.CreateDeployment(deploymentConf)

	if err = d.executor.ApplyAndWait(deployment, timeout); err != nil {
		return err
	}

	serviceConf, err := d.configMapper.GetServiceConfig(resource.GRPCServerService)
	if err != nil {
		return err
	}

	service := d.resourceFactory.CreateService(serviceConf)

	return d.executor.ApplyAndWait(service, timeout)
}

func (d *DeploymentManager) DeployWorkerpools() error {
	pvConf, err := d.configMapper.GetPersistentVolumeConfig(resource.WorkerpoolPV)
	if err != nil {
		return err
	}

	pv := d.resourceFactory.CreatePersistentVolume(pvConf)

	if err = d.executor.ApplyAndWait(pv, timeout); err != nil {
		return err
	}

	pvcConf, err := d.configMapper.GetPersistentVolumeClaimConfig(resource.WorkerpoolPVC)
	if err != nil {
		return err
	}

	pvc := d.resourceFactory.CreatePersistentVolumeClaim(pvcConf)

	if err = d.executor.ApplyAndWait(pvc, timeout); err != nil {
		return err
	}

	deploymentConf, err := d.configMapper.GetDeploymentConfig(resource.WorkerpoolDeployment)
	if err != nil {
		return err
	}

	deployment := d.resourceFactory.CreateDeployment(deploymentConf)

	return d.executor.ApplyAndWait(deployment, timeout)
}

func (d *DeploymentManager) DestroyAll() error {
	// Deleting the namespace will delete all associated resources
	namespaceConf, err := d.configMapper.GetNamespaceConfig(resource.Namespace)
	if err != nil {
		return err
	}

	namespace := d.resourceFactory.CreateNamespace(namespaceConf)

	if err = d.executor.DestroyAndWait(namespace, timeout); err != nil {
		return err
	}

	// Persistent volumes are not scoped to a namespace
	pvConf, err := d.configMapper.GetPersistentVolumeConfig(resource.WorkerpoolPV)
	if err != nil {
		return err
	}

	pv := d.resourceFactory.CreatePersistentVolume(pvConf)

	if err = d.executor.DestroyAndWait(pv, timeout); err != nil {
		return err
	}

	d.logger.Success("GoFlow destroyed!")

	return nil
}
