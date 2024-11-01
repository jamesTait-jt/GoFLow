package resource

import (
	"context"
	"fmt"

	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/pkg/slice"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type Waiter struct {
	ctx    context.Context
	logger log.Logger
}

type Watchable interface {
	Watch(ctx context.Context, options metav1.ListOptions) (watch.Interface, error)
}

func NewWaiter(ctx context.Context, logger log.Logger) *Waiter {
	return &Waiter{ctx: ctx, logger: logger}
}

func (w *Waiter) WaitFor(
	resourceName, namespace string,
	eventTypes []watch.EventType,
	client Watchable,
) error {
	stopLog := w.logger.Waiting(fmt.Sprintf("Waiting for event of type %v from '%s'", eventTypes, resourceName))

	// This won't be populated if we're destroying or creating a namespace
	var fieldSelector string
	if namespace != "" {
		fieldSelector = fmt.Sprintf("metadata.name=%s,metadata.namespace=%s", resourceName, namespace)
	} else {
		fieldSelector = fmt.Sprintf("metadata.name=%s", resourceName)
	}

	watcher, err := client.Watch(w.ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		stopLog("Failed to watch resource", false)
		return err
	}
	defer watcher.Stop()

	for event := range watcher.ResultChan() {
		if slice.Contains(eventTypes, event.Type) {
			w.logger.Info(fmt.Sprintf("%v", event))
			stopLog(fmt.Sprintf("Found event of type '%v'", event.Type), true)

			return nil
		}
	}

	// TODO: This will never reach as we don't have a timeout
	return nil
}
