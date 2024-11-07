package taskhandlers

import (
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/workerpool/plugin"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
)

func Load(pluginDir string) (*store.InMemoryKVStore[string, task.Handler], error) {
	plugins, err := plugin.Load(pluginDir)
	if err != nil {
		return nil, err
	}

	taskHandlers := store.NewInMemoryKVStore[string, task.Handler]()

	for pluginName, plg := range plugins {
		symbol, err := plg.Lookup("NewHandler")
		if err != nil {
			return nil, err
		}

		handlerFactory, ok := symbol.(func() task.Handler)
		if !ok {
			return nil, fmt.Errorf("invalid plugin: Handler does not implement Handler interface")
		}

		handler := handlerFactory()

		taskHandlers.Put(pluginName, handler)
	}

	return taskHandlers, nil
}
