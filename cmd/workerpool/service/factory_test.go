//go:build unit

package service

import (
	"testing"

	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/pkg/serialise"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func Test_WorkerpoolFactory_NewRedisWorkerpoolService(t *testing.T) {
	t.Run("Initialises a workerpool service with redis backed brokers", func(t *testing.T) {
		// Arrange
		pool := new(mockWorkerpoolRunner)
		resultSerialiser := serialise.NewGobSerialiser[task.Result]()
		taskSerialiser := serialise.NewGobSerialiser[task.Task]()
		taskHandlers := store.NewInMemoryKVStore[string, task.Handler]()
		logger := log.NewConsoleLogger()

		client := &redis.Client{}

		f := NewFactory(pool, resultSerialiser, taskSerialiser, taskHandlers, logger)

		// Act
		service := f.CreateRedisWorkerpoolService(client)

		// Assert
		assert.NotNil(t, service)
		assert.Equal(t, pool, service.pool)
		assert.Equal(t, taskHandlers, service.taskHandlers)
		assert.Implements(t, (*task.Dequeuer[task.Task])(nil), service.taskQueue)
		assert.Implements(t, (*task.Submitter[task.Result])(nil), service.resultQueue)
	})
}
