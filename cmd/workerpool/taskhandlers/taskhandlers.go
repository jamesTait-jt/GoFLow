package taskhandlers

import (
	"fmt"
	"plugin"

	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
)

type symbolFinder interface {
	Lookup(symName string) (plugin.Symbol, error)
}

type pluginLoader interface {
	Load(pluginDir string) (map[string]symbolFinder, error)
}

func Load(pluginLoader pluginLoader, pluginDir string) (*store.InMemoryKVStore[string, task.Handler], error) {
	plugins, err := pluginLoader.Load(pluginDir)
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
