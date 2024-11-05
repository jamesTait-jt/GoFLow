package k8s

import (
	"fmt"
	"time"

	"github.com/jamesTait-jt/goflow/pkg/log"
	"k8s.io/apimachinery/pkg/watch"
)

type operator interface {
	Apply(kubeResource ApplyGetter) (bool, error)
	Delete(kubeResource Deleter) (bool, error)
	WaitFor(kubeResource Watcher, eventTypes []watch.EventType, timeout time.Duration) error
}

type identifier interface {
	Name() string
	Kind() string
}

type identifiableWatchableApplyGetter interface {
	identifier
	ApplyGetter
	Watcher
}

type identifiableWatchableDeleter interface {
	identifier
	Deleter
	Watcher
}

type DeploymentExecutor struct {
	op     operator
	logger log.Logger
}

func NewDeploymentExecutor(logger log.Logger) *DeploymentExecutor {
	return &DeploymentExecutor{
		op:     NewOperator(),
		logger: logger,
	}
}

func (d *DeploymentExecutor) ApplyAndWait(kubeResource identifiableWatchableApplyGetter, timeout time.Duration) error {
	d.logger.Info(fmt.Sprintf("Deploying %s '%s'", kubeResource.Kind(), kubeResource.Name()))

	neededModification, err := d.op.Apply(kubeResource)
	if err != nil {
		return err
	}

	if !neededModification {
		d.logger.Success(fmt.Sprintf("'%s' deployed successfully", kubeResource.Name()))

		return nil
	}

	d.logger.Info(fmt.Sprintf("'%s' needs modification - applying changes", kubeResource.Name()))

	if err := d.op.WaitFor(kubeResource, []watch.EventType{watch.Added, watch.Modified}, timeout); err != nil {
		return err
	}

	d.logger.Success(fmt.Sprintf("'%s' deployed successfully", kubeResource.Name()))

	return nil
}

func (d *DeploymentExecutor) DestroyAndWait(kubeResource identifiableWatchableDeleter, timeout time.Duration) error {
	d.logger.Info(fmt.Sprintf("Destroying %s '%s'", kubeResource.Kind(), kubeResource.Name()))

	neededDeleting, err := d.op.Delete(kubeResource)
	if err != nil {
		return err
	}

	if !neededDeleting {
		d.logger.Warn(fmt.Sprintf("couldnt find '%s'", kubeResource.Name()))

		return nil
	}

	d.logger.Info(fmt.Sprintf("'%s' needs destroying - waiting...", kubeResource.Name()))

	if err := d.op.WaitFor(kubeResource, []watch.EventType{watch.Deleted}, timeout); err != nil {
		return err
	}

	d.logger.Success(fmt.Sprintf("'%s' destroyed successfully", kubeResource.Name()))

	return nil
}
