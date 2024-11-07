package pluginloader

import (
	"fmt"
	"os"
	"plugin"
	"strings"
)

type SymbolFinder interface {
	Lookup(symName string) (plugin.Symbol, error)
}

type Loader struct{}

func (l *Loader) Load(pluginDir string) (map[string]SymbolFinder, error) {
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, err
	}

	plugins := make(map[string]SymbolFinder, len(files))

	for i := 0; i < len(files); i++ {
		plg, err := plugin.Open(fmt.Sprintf("%s/%s", pluginDir, files[i].Name()))
		if err != nil {
			return nil, err
		}

		fmt.Printf("%s/%s\n", pluginDir, files[i].Name())

		pluginName := strings.TrimSuffix(files[i].Name(), ".so")

		plugins[pluginName] = plg
	}

	return plugins, nil
}
