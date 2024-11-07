package taskhandlers

import (
	"errors"
	"plugin"
	"testing"

	"github.com/jamesTait-jt/goflow/cmd/workerpool/pluginloader"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Load(t *testing.T) {
	t.Run("Loads multiple plugins and returns a store with their handlers", func(t *testing.T) {
		// Arrange
		pluginLoader := new(mockPluginLoader)
		pluginDir := "plugin-dir"

		keyOne := "keyOne"
		keyTwo := "keyTwo"
		pluginOne := &mockSymbolFinder{}
		pluginTwo := &mockSymbolFinder{}

		pluginLoader.On("Load", pluginDir).Once().Return(
			map[string]pluginloader.SymbolFinder{
				keyOne: pluginOne,
				keyTwo: pluginTwo,
			},
			nil,
		)

		resultOne := task.Result{TaskID: "1"}
		var handlerOne task.Handler = func(payload any) task.Result { return resultOne }
		symbolOne := func() task.Handler { return handlerOne }
		pluginOne.On("Lookup", "NewHandler").Once().Return(symbolOne, nil)

		resultTwo := task.Result{TaskID: "2"}
		var handlerTwo task.Handler = func(payload any) task.Result { return resultTwo }
		symbolTwo := func() task.Handler { return handlerTwo }
		pluginTwo.On("Lookup", "NewHandler").Once().Return(symbolTwo, nil)

		// Act
		handlers, err := Load(pluginLoader, pluginDir)

		// Assert
		assert.NoError(t, err)

		returnedHandlerOne, ok := handlers.Get(keyOne)
		assert.True(t, ok)
		assert.Equal(t, resultOne, returnedHandlerOne(nil))

		returnedHandlerTwo, ok := handlers.Get(keyTwo)
		assert.True(t, ok)
		assert.Equal(t, resultTwo, returnedHandlerTwo(nil))
	})

	t.Run("Returns an error if could not load the plugins", func(t *testing.T) {
		// Arrange
		pluginLoader := new(mockPluginLoader)
		pluginDir := "plugin-dir"

		loadErr := errors.New("couldn't load plugins")
		pluginLoader.On("Load", pluginDir).Once().Return(nil, loadErr)

		// Act
		handlers, err := Load(pluginLoader, pluginDir)

		// Assert
		assert.EqualError(t, err, loadErr.Error())
		assert.Nil(t, handlers)
	})

	t.Run("Returns an error if symbol lookup fails", func(t *testing.T) {
		// Arrange
		pluginLoader := new(mockPluginLoader)
		pluginDir := "plugin-dir"

		keyOne := "keyOne"
		pluginOne := &mockSymbolFinder{}

		pluginLoader.On("Load", pluginDir).Once().Return(
			map[string]pluginloader.SymbolFinder{
				keyOne: pluginOne,
			},
			nil,
		)

		lookupErr := errors.New("couldn't lookup symbol")
		pluginOne.On("Lookup", "NewHandler").Once().Return(nil, lookupErr)

		// Act
		handlers, err := Load(pluginLoader, pluginDir)

		// Assert
		assert.EqualError(t, err, lookupErr.Error())
		assert.Nil(t, handlers)
	})

	t.Run("Returns an error if symbol couldn't be cast to task.Handler", func(t *testing.T) {
		// Arrange
		pluginLoader := new(mockPluginLoader)
		pluginDir := "plugin-dir"

		keyOne := "keyOne"
		pluginOne := &mockSymbolFinder{}

		pluginLoader.On("Load", pluginDir).Once().Return(
			map[string]pluginloader.SymbolFinder{
				keyOne: pluginOne,
			},
			nil,
		)

		handlerOne := func() any { return nil }
		symbolOne := func() any { return handlerOne }
		pluginOne.On("Lookup", "NewHandler").Once().Return(symbolOne, nil)

		// Act
		handlers, err := Load(pluginLoader, pluginDir)

		// Assert
		assert.EqualError(t, err, "invalid plugin: Handler does not implement Handler interface")
		assert.Nil(t, handlers)
	})
}

type mockSymbolFinder struct {
	mock.Mock
}

func (m *mockSymbolFinder) Lookup(symName string) (plugin.Symbol, error) {
	args := m.Called(symName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(plugin.Symbol), args.Error(1)
}

type mockPluginLoader struct {
	mock.Mock
}

func (m *mockPluginLoader) Load(pluginDir string) (map[string]pluginloader.SymbolFinder, error) {
	args := m.Called(pluginDir)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]pluginloader.SymbolFinder), args.Error(1)
}
