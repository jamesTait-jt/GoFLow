package pluginloader

import (
	"errors"
	"plugin"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func Test_Loader_Load(t *testing.T) {
	t.Run("Loads plugins successfully from a directory", func(t *testing.T) {
		// Arrange
		fs := afero.NewMemMapFs()

		dirName := "plugins"
		filepathOne := "plugin1.so"
		filepathTwo := "plugin2.so"
		fs.Create(dirName + "/" + filepathOne)
		fs.Create(dirName + "/" + filepathTwo)

		pluginOne := new(plugin.Plugin)
		pluginTwo := new(plugin.Plugin)

		var openPlugin pluginOpener = func(path string) (*plugin.Plugin, error) {
			switch path {
			case dirName + "/" + filepathOne:
				return pluginOne, nil
			case dirName + "/" + filepathTwo:
				return pluginTwo, nil
			default:
				return nil, errors.New("wrong path")
			}
		}

		loader := New(fs, openPlugin)

		// Act
		plugins, err := loader.Load(dirName)

		// Assert
		assert.NoError(t, err)

		assert.Len(t, plugins, 2)
		assert.Equal(t, pluginOne, plugins["plugin1"])
		assert.Equal(t, pluginTwo, plugins["plugin2"])
	})

	t.Run("Returns error if couldn't open directory", func(t *testing.T) {
		// Arrange
		fs := afero.NewMemMapFs()

		loader := New(fs, nil)

		// Act
		plugins, err := loader.Load("doesnt exist")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, plugins)
	})

	t.Run("Returns error if couldn't open plugin", func(t *testing.T) {
		// Arrange
		fs := afero.NewMemMapFs()

		dirName := "plugins"
		filepathOne := "plugin1.so"
		fs.Create(dirName + "/" + filepathOne)

		openPluginErr := errors.New("couldnt open plugin")
		var openPlugin pluginOpener = func(path string) (*plugin.Plugin, error) {
			return nil, openPluginErr
		}

		loader := New(fs, openPlugin)

		// Act
		plugins, err := loader.Load(dirName)

		// Assert
		assert.EqualError(t, err, openPluginErr.Error())
		assert.Nil(t, plugins)
	})
}
