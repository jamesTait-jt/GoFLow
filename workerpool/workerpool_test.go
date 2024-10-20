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
			Error:   nil,
			Payload: "PayloadData",
		}

		handlerCalled := false
		taskHandlers.Put(taskType, func(payload any) task.Result {
			handlerCalled = true

			return resultToReturn
		})

		// Act
		wp.Start(ctx, taskQueue, resultQueue, taskHandlers)
		taskQueue.Submit(ctx, submittedTask)
		receivedResult := <-resultQueue.Dequeue(ctx)

		// Assert
		assert.True(t, handlerCalled)
		assert.Equal(t, submittedTask.ID, resultToReturn.TaskID)
		assert.Equal(t, resultToReturn.Payload, receivedResult.Payload)
		assert.Nil(t, receivedResult.Error)

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
			Error:   errors.New("error"),
			Payload: nil,
		}

		handlerCalled := false
		taskHandlers.Put(taskType, func(payload any) task.Result {
			handlerCalled = true

			return resultToReturn
		})

		// Act
		wp.Start(ctx, taskQueue, resultQueue, taskHandlers)
		taskQueue.Submit(ctx, submittedTask)
		receivedResult := <-resultQueue.Dequeue(ctx)

		// Assert
		assert.True(t, handlerCalled)
		assert.Equal(t, submittedTask.ID, resultToReturn.TaskID)
		assert.EqualError(t, resultToReturn.Error, receivedResult.Error.Error())
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
			Error:   nil,
			Payload: "PayloadData",
		}

		handlerCalled := false
		taskHandlers.Put("not a key", func(payload any) task.Result {
			handlerCalled = true

			return resultToReturn
		})

		// Act
		wp.Start(ctx, taskQueue, resultQueue, taskHandlers)
		taskQueue.Submit(ctx, submittedTask)

		// Assert
		assert.False(t, handlerCalled)

		cancel()
		wg.Wait()
	})
}

// func TestPool_Start(t *testing.T) {
// 	t.Run("Starts all workers in the pool", func(t *testing.T) {
// 		// Arrange
// 		numWorkers := 5
// 		ctx := context.Background()
// 		wg := &sync.WaitGroup{}
// 		taskSource := &mockTaskSource{}
// 		resultsCh := make(chan<- task.Result)

// 		mockWorkers := make(map[int]*mockWorker)

// 		for i := 0; i < numWorkers; i++ {
// 			mockWorker := new(mockWorker)
// 			mockWorker.On("Start", ctx, wg, taskSource, resultsCh).Once()
// 			mockWorkers[i] = mockWorker
// 		}

// 		pool := &Pool{
// 			workers: make(map[string]Worker),
// 			wg:      wg,
// 		}

// 		for i := 0; i < numWorkers; i++ {
// 			pool.workers[fmt.Sprintf("%d", i)] = mockWorkers[i]
// 		}

// 		// Act
// 		pool.Start(ctx, taskSource, resultsCh)

// 		// Assert
// 		for i := 0; i < numWorkers; i++ {
// 			mockWorkers[i].AssertCalled(t, "Start", ctx, wg, taskSource, resultsCh)
// 		}
// 	})
// }
