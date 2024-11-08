//go:build unit

package workerpool

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("Creates a new worker pool with variables initialised", func(t *testing.T) {
		// Arrange
		numWorkers := 5

		// Act
		wp := New(numWorkers)

		// Assert
		assert.Equal(t, numWorkers, wp.numWorkers)
		assert.NotNil(t, wp.wg)
	})
}

func Test_Pool_Start(t *testing.T) {
	t.Run("Starts the workers and processes task", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())

		taskQueue := broker.NewChannelBroker[task.Task](0)
		resultQueue := broker.NewChannelBroker[task.Result](0)
		taskHandlers := store.NewInMemoryKVStore[string, task.Handler]()

		wg := &sync.WaitGroup{}
		wp := Pool{numWorkers: 1, wg: wg}

		taskType := "test_task"

		submittedTask := task.Task{Type: taskType}

		resultToReturn := task.Result{
			ErrMsg:  "",
			Payload: "PayloadData",
		}

		handlerCalled := false

		taskHandlers.Put(taskType, func(_ any) task.Result {
			handlerCalled = true

			return resultToReturn
		})

		// Act
		wp.Start(ctx, taskQueue, resultQueue, taskHandlers)
		_ = taskQueue.Submit(ctx, submittedTask)

		receivedResult := <-resultQueue.Dequeue(ctx)

		// Assert
		assert.True(t, handlerCalled)
		assert.Equal(t, submittedTask.ID, resultToReturn.TaskID)
		assert.Equal(t, resultToReturn.Payload, receivedResult.Payload)
		assert.Equal(t, "", receivedResult.ErrMsg)

		cancel()
		wg.Wait()
	})

	t.Run("Starts the workers and processes task even with errors", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())

		taskQueue := broker.NewChannelBroker[task.Task](0)
		resultQueue := broker.NewChannelBroker[task.Result](0)
		taskHandlers := store.NewInMemoryKVStore[string, task.Handler]()

		wg := &sync.WaitGroup{}
		wp := Pool{numWorkers: 1, wg: wg}

		taskType := "test_task"

		submittedTask := task.Task{Type: taskType}

		resultToReturn := task.Result{
			ErrMsg:  errors.New("error").Error(),
			Payload: nil,
		}

		handlerCalled := false

		taskHandlers.Put(taskType, func(_ any) task.Result {
			handlerCalled = true

			return resultToReturn
		})

		// Act
		wp.Start(ctx, taskQueue, resultQueue, taskHandlers)
		_ = taskQueue.Submit(ctx, submittedTask)
		receivedResult := <-resultQueue.Dequeue(ctx)

		// Assert
		assert.True(t, handlerCalled)
		assert.Equal(t, submittedTask.ID, resultToReturn.TaskID)
		assert.Equal(t, resultToReturn.ErrMsg, receivedResult.ErrMsg)
		assert.Nil(t, receivedResult.Payload)

		cancel()
		wg.Wait()
	})

	t.Run("Skips task if handler not registered for type", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())

		taskQueue := broker.NewChannelBroker[task.Task](0)
		resultQueue := broker.NewChannelBroker[task.Result](0)
		taskHandlers := store.NewInMemoryKVStore[string, task.Handler]()

		wg := &sync.WaitGroup{}
		wp := Pool{numWorkers: 1, wg: wg}

		taskType := "test_task"
		submittedTask := task.Task{Type: taskType}

		resultToReturn := task.Result{
			ErrMsg:  "",
			Payload: "PayloadData",
		}

		handlerCalled := false

		taskHandlers.Put("not a key", func(_ any) task.Result {
			handlerCalled = true

			return resultToReturn
		})

		// Act
		wp.Start(ctx, taskQueue, resultQueue, taskHandlers)
		_ = taskQueue.Submit(ctx, submittedTask)

		// Assert
		assert.False(t, handlerCalled)

		cancel()
		wg.Wait()
	})
}
