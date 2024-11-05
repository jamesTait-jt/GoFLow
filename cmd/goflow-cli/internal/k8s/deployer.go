package k8s

import (
	"fmt"
	"time"

	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/grpcserver"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/redis"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/resource"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s/workerpool"
	"github.com/jamesTait-jt/goflow/pkg/log"

	"k8s.io/apimachinery/pkg/watch"
	acapiv1 "k8s.io/client-go/applyconfigurations/core/v1"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedapiv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var timeout = 30 * time.Second

type operator interface {
	Apply(kubeResource ApplyGetter) (bool, error)
	Delete(kubeResource Deleter) (bool, error)
	WaitFor(kubeResource Watcher, eventTypes []watch.EventType, timeout time.Duration) error
}

type clientset interface {
	Namespaces() typedapiv1.NamespaceInterface
	Deployments() typedappsv1.DeploymentInterface
	Services() typedapiv1.ServiceInterface
	PersistentVolumes() typedapiv1.PersistentVolumeInterface
	PersistentVolumeClaims() typedapiv1.PersistentVolumeClaimInterface
}

type Deployer struct {
	conf      *config.Config
	logger    log.Logger
	op        operator
	clientset clientset
}

func NewDeployer(conf *config.Config, logger log.Logger, clients clientset, op operator) (*Deployer, error) {
	return &Deployer{
		conf:      conf,
		logger:    logger,
		op:        op,
		clientset: clients,
	}, nil
}

func (d *Deployer) DeployNamespace() error {
	namespacesClient := d.clientset.Namespaces()

	namespace := resource.NewNamespace(acapiv1.Namespace(d.conf.Kubernetes.Namespace), namespacesClient)

	return applyAndWait(namespace, d.op, d.logger, timeout)
}

func (d *Deployer) DeployMessageBroker() error {
	deploymentsClient := d.clientset.Deployments()
	servicesClient := d.clientset.Services()

	deployment := resource.NewDeployment(redis.Deployment(d.conf), deploymentsClient)
	if err := applyAndWait(deployment, d.op, d.logger, timeout); err != nil {
		return err
	}

	service := resource.NewService(redis.Service(d.conf), servicesClient)

	return applyAndWait(service, d.op, d.logger, timeout)
}

func (d *Deployer) DeployGRPCServer() error {
	deploymentsClient := d.clientset.Deployments()
	servicesClient := d.clientset.Services()

	deployment := resource.NewDeployment(grpcserver.Deployment(d.conf), deploymentsClient)
	if err := applyAndWait(deployment, d.op, d.logger, timeout); err != nil {
		return err
	}

	service := resource.NewService(grpcserver.Service(d.conf), servicesClient)

	return applyAndWait(service, d.op, d.logger, timeout)
}

func (d *Deployer) DeployWorkerpools() error {
	pvClient := d.clientset.PersistentVolumes()
	pvcClient := d.clientset.PersistentVolumeClaims()
	deploymentsClient := d.clientset.Deployments()

	pv := resource.NewPersistentVolume(workerpool.HandlersPV(d.conf), pvClient)
	if err := applyAndWait(pv, d.op, d.logger, timeout); err != nil {
		return err
	}

	pvc := resource.NewPersistentVolumeClaim(workerpool.HandlersPVC(d.conf), pvcClient)
	if err := applyAndWait(pvc, d.op, d.logger, timeout); err != nil {
		return err
	}

	deployment := resource.NewDeployment(workerpool.Deployment(d.conf), deploymentsClient)

	return applyAndWait(deployment, d.op, d.logger, timeout)
}

func (d *Deployer) DestroyAll() error {
	namespacesClient := d.clientset.Namespaces()
	pvClient := d.clientset.PersistentVolumes()

	// Deleting the namespace will delete all associated resources
	namespace := resource.NewNamespace(acapiv1.Namespace(d.conf.Kubernetes.Namespace), namespacesClient)
	if err := destroyAndWait(namespace, d.op, d.logger); err != nil {
		return err
	}

	// Persistent volumes are not scoped to a namespace
	pv := resource.NewPersistentVolume(workerpool.HandlersPV(d.conf), pvClient)
	if err := destroyAndWait(pv, d.op, d.logger); err != nil {
		return err
	}

	d.logger.Success("GoFlow destroyed!")

	return nil
}

type Identifier interface {
	Name() string
	Kind() string
}

type IdentifiableWatchableApplyGetter interface {
	Identifier
	ApplyGetter
	Watcher
}

func applyAndWait(
	kubeResource IdentifiableWatchableApplyGetter,
	kubeOperator operator,
	logger log.Logger,
	deployTimeout time.Duration,
) error {
	logger.Info(fmt.Sprintf("Deploying %s '%s'", kubeResource.Kind(), kubeResource.Name()))

	neededModification, err := kubeOperator.Apply(kubeResource)
	if err != nil {
		return err
	}

	if !neededModification {
		logger.Success(fmt.Sprintf("'%s' deployed successfully", kubeResource.Name()))

		return nil
	}

	logger.Info(fmt.Sprintf("'%s' needs modification - applying changes", kubeResource.Name()))

	if err := kubeOperator.WaitFor(kubeResource, []watch.EventType{watch.Added, watch.Modified}, deployTimeout); err != nil {
		return err
	}

	logger.Success(fmt.Sprintf("'%s' deployed successfully", kubeResource.Name()))

	return nil
}

type IdentifiableWatchableDeleter interface {
	Identifier
	Deleter
	Watcher
}

func destroyAndWait(kubeResource IdentifiableWatchableDeleter, kubeOperator operator, logger log.Logger) error {
	logger.Info(fmt.Sprintf("Destroying %s '%s'", kubeResource.Kind(), kubeResource.Name()))

	neededDeleting, err := kubeOperator.Delete(kubeResource)
	if err != nil {
		return err
	}

	if !neededDeleting {
		logger.Warn(fmt.Sprintf("couldnt find '%s'", kubeResource.Name()))

		return nil
	}

	logger.Info(fmt.Sprintf("'%s' needs destroying - waiting...", kubeResource.Name()))

	if err := kubeOperator.WaitFor(kubeResource, []watch.EventType{watch.Deleted}, timeout); err != nil {
		return err
	}

	logger.Success(fmt.Sprintf("'%s' destroyed successfully", kubeResource.Name()))

	return nil
}
