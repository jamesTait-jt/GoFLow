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

type Manager struct {
	conf      *config.Config
	logger    log.Logger
	op        operator
	clientset clientset
}

func NewManager(conf *config.Config, logger log.Logger, clients clientset, op operator) (*Manager, error) {
	return &Manager{
		conf:      conf,
		logger:    logger,
		op:        op,
		clientset: clients,
	}, nil
}

func (m *Manager) Deploy() error {
	namespacesClient := m.clientset.Namespaces()
	deploymentsClient := m.clientset.Deployments()
	servicesClient := m.clientset.Services()
	pvClient := m.clientset.PersistentVolumes()
	pvcClient := m.clientset.PersistentVolumeClaims()

	namespace := resource.NewNamespace(acapiv1.Namespace(m.conf.Kubernetes.Namespace), namespacesClient)
	if err := applyAndWait(namespace, m.op, m.logger, timeout); err != nil {
		return fmt.Errorf("failed to deploy namespace: %w", err)
	}

	if err := deployMessageBroker(m.conf, m.op, deploymentsClient, servicesClient, m.logger, timeout); err != nil {
		return fmt.Errorf("failed to deploy message broker: %w", err)
	}

	if err := deployGRPCServer(m.conf, m.op, deploymentsClient, servicesClient, m.logger, timeout); err != nil {
		return fmt.Errorf("failed to deploy gRPC server: %w", err)
	}

	if err := deployWorkerpools(m.conf, m.op, deploymentsClient, pvClient, pvcClient, m.logger, timeout); err != nil {
		return fmt.Errorf("failed to deploy worker pools: %w", err)
	}

	m.logger.Success(
		fmt.Sprintf(
			"GoFlow deployed! Use `kubectl get pods -n %s` to see the status of the application",
			m.conf.Kubernetes.Namespace,
		),
	)

	return nil
}

func (m *Manager) Destroy() error {
	namespacesClient := m.clientset.Namespaces()
	pvClient := m.clientset.PersistentVolumes()

	// Deleting the namespace will delete all associated resources
	namespace := resource.NewNamespace(acapiv1.Namespace(m.conf.Kubernetes.Namespace), namespacesClient)
	if err := destroyAndWait(namespace, m.op, m.logger); err != nil {
		return err
	}

	// Persistent volumes are not scoped to a namespace
	pv := resource.NewPersistentVolume(workerpool.HandlersPV(m.conf), pvClient)
	if err := destroyAndWait(pv, m.op, m.logger); err != nil {
		return err
	}

	m.logger.Success("GoFlow destroyed!")

	return nil
}

func deployMessageBroker(
	conf *config.Config,
	kubeOperator operator,
	deploymentsClient resource.DeploymentInterface,
	servicesClient resource.ServicesInterface,
	logger log.Logger,
	timeout time.Duration,
) error {
	deployment := resource.NewDeployment(redis.Deployment(conf), deploymentsClient)
	if err := applyAndWait(deployment, kubeOperator, logger, timeout); err != nil {
		return err
	}

	service := resource.NewService(redis.Service(conf), servicesClient)

	return applyAndWait(service, kubeOperator, logger, timeout)
}

func deployGRPCServer(
	conf *config.Config,
	kubeOperator operator,
	deploymentsClient resource.DeploymentInterface,
	servicesClient resource.ServicesInterface,
	logger log.Logger,
	timeout time.Duration,
) error {
	deployment := resource.NewDeployment(grpcserver.Deployment(conf), deploymentsClient)
	if err := applyAndWait(deployment, kubeOperator, logger, timeout); err != nil {
		return err
	}

	service := resource.NewService(grpcserver.Service(conf), servicesClient)

	return applyAndWait(service, kubeOperator, logger, timeout)
}

func deployWorkerpools(
	conf *config.Config,
	kubeOperator operator,
	deploymentsClient resource.DeploymentInterface,
	pvClient resource.PersistentVolumeInterface,
	pvcClient resource.PersistentVolumeClaimInterface,
	logger log.Logger,
	timeout time.Duration,
) error {
	pv := resource.NewPersistentVolume(workerpool.HandlersPV(conf), pvClient)
	if err := applyAndWait(pv, kubeOperator, logger, timeout); err != nil {
		return err
	}

	pvc := resource.NewPersistentVolumeClaim(workerpool.HandlersPVC(conf), pvcClient)
	if err := applyAndWait(pvc, kubeOperator, logger, timeout); err != nil {
		return err
	}

	deployment := resource.NewDeployment(workerpool.Deployment(conf), deploymentsClient)

	return applyAndWait(deployment, kubeOperator, logger, timeout)
}

type IdentifiableWatchableApplyGetter interface {
	Name() string
	Kind() string
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
	Name() string
	Kind() string
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
