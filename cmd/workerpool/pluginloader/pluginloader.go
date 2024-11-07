package pluginloader

import (
	"fmt"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/spf13/afero"
)

type SymbolFinder interface {
	Lookup(symName string) (plugin.Symbol, error)
}

type pluginOpener func(path string) (*plugin.Plugin, error)

type Loader struct {
	fs         afero.Fs
	openPlugin pluginOpener
}

func New(fs afero.Fs, openPlugin pluginOpener) *Loader {
	return &Loader{fs: fs, openPlugin: openPlugin}
}

func (l *Loader) Load(pluginDir string) (map[string]SymbolFinder, error) {
	files, err := afero.ReadDir(l.fs, pluginDir)
	if err != nil {
		return nil, err
	}

	plugins := make(map[string]SymbolFinder, len(files))

	for i := 0; i < len(files); i++ {
		// Only process files with ".so" extension
		if filepath.Ext(files[i].Name()) != ".so" {
			continue
		}

		plg, err := l.openPlugin(fmt.Sprintf("%s/%s", pluginDir, files[i].Name()))
		if err != nil {
			return nil, err
		}

		fmt.Printf("%s/%s\n", pluginDir, files[i].Name())

		pluginName := strings.TrimSuffix(files[i].Name(), ".so")

		plugins[pluginName] = plg
	}

	return plugins, nil
}
