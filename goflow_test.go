//go:build unit

package goflow

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/pkg/channel"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/jamesTait-jt/goflow/workerpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_New(t *testing.T) {
	t.Run("Initialises goflow with default options in distributed mode", func(t *testing.T) {
		// Arrange
		taskBroker := broker.NewChannelBroker[task.Task](1)
		resultBroker := broker.NewChannelBroker[task.Result](1)

		// Act
		gf := New(taskBroker, resultBroker)

		// Assert
		assert.NotNil(t, gf)
		assert.NotNil(t, gf.ctx)
		assert.NotNil(t, gf.cancel)
		assert.NotNil(t, gf.resultsWriterWG)

		assert.Nil(t, gf.workers)
		assert.Nil(t, gf.taskHandlers)

		assert.Equal(t, taskBroker, gf.taskBroker)
		assert.Equal(t, resultBroker, gf.resultsBroker)

		assert.IsType(t, &store.InMemoryKVStore[string, task.Result]{}, gf.results)

		assert.False(t, gf.started)
	})

	t.Run("Initialises goflow with custom options in distributed mode", func(t *testing.T) {
		// Arrange
		resultStore := store.NewInMemoryKVStore[string, task.Result]()

		// Act
		gf := New(
			nil,
			nil,
			WithResultsStore(resultStore),
		)

		// Assert
		assert.Equal(t, resultStore, gf.results)
	})
}

func Test_NewLocalMode(t *testing.T) {
	t.Run("Initialises goflow with default options in local mode", func(t *testing.T) {
		// Arrange
		taskHandlers := store.NewInMemoryKVStore[string, task.Handler]()

		// Act
		gf := NewLocalMode(taskHandlers)

		// Assert
		assert.NotNil(t, gf)
		assert.NotNil(t, gf.ctx)
		assert.NotNil(t, gf.cancel)
		assert.NotNil(t, gf.resultsWriterWG)

		assert.IsType(t, &workerpool.Pool{}, gf.workers)
		assert.IsType(t, &broker.ChannelBroker[task.Task]{}, gf.taskBroker)
		assert.IsType(t, &broker.ChannelBroker[task.Result]{}, gf.resultsBroker)
		assert.IsType(t, &store.InMemoryKVStore[string, task.Result]{}, gf.results)

		assert.Equal(t, taskHandlers, gf.taskHandlers)

		assert.False(t, gf.started)
	})

	t.Run("Initialises goflow with custom options in local mode", func(t *testing.T) {
		// Arrange
		resultStore := store.NewInMemoryKVStore[string, task.Result]()

		// Act
		gf := NewLocalMode(
			nil,
			WithNumWorkers(10),
			WithTaskQueueBufferSize(10),
			WithResultQueueBufferSize(10),
			WithResultsStore(resultStore),
		)

		// Assert
		assert.IsType(t, &workerpool.Pool{}, gf.workers)
		assert.IsType(t, &broker.ChannelBroker[task.Task]{}, gf.taskBroker)
		assert.IsType(t, &broker.ChannelBroker[task.Result]{}, gf.resultsBroker)
		assert.Equal(t, resultStore, gf.results)
	})
}

func Test_GoFlow_Start(t *testing.T) {
	t.Run("Returns error if GoFlow already started", func(t *testing.T) {
		// Arrange
		gf := &GoFlow{
			started: true,
		}

		// Act
		err := gf.Start()

		// Assert
		assert.EqualError(t, err, ErrAlreadyStarted.Error())
	})

	t.Run("Does not start the workerpool if workers not initialised", func(_ *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		resultsWriterWG := &sync.WaitGroup{}

		taskHandlers := store.NewInMemoryKVStore[string, task.Handler]()
		gf := &GoFlow{
			ctx:             ctx,
			cancel:          cancel,
			workers:         nil,
			taskHandlers:    taskHandlers,
			resultsBroker:   broker.NewChannelBroker[task.Result](0),
			resultsWriterWG: resultsWriterWG,
			started:         false,
		}

		// Act
		cancel()
		err := gf.Start()

		// Assert - it's not really possible to assert here but there would be a nil
		// pointer dereference if Start() were called on the nil workers, so we can
		// assume a pass if there is no panic
		resultsWriterWG.Wait()
		assert.Nil(t, err)
		assert.True(t, gf.started)
	})

	t.Run("Does not start the workerpool if task handlers not initialised", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		resultsWriterWG := &sync.WaitGroup{}

		workers := new(mockWorkerPool)
		gf := &GoFlow{
			ctx:             ctx,
			cancel:          cancel,
			workers:         workers,
			taskHandlers:    nil,
			resultsBroker:   broker.NewChannelBroker[task.Result](0),
			resultsWriterWG: resultsWriterWG,
			started:         false,
		}

		// Act
		cancel()
		err := gf.Start()

		// Assert
		resultsWriterWG.Wait()
		assert.Nil(t, err)
		workers.AssertNotCalled(t, "Start")
		assert.True(t, gf.started)
	})

	t.Run("Starts the workerpool and persists incoming results", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		resultsWriterWG := &sync.WaitGroup{}

		taskBroker := new(mockBroker[task.Task])
		resultBroker := new(mockBroker[task.Result])
		resultStore := new(mockKVStore[string, task.Result])
		taskHandlers := new(mockKVStore[string, task.Handler])
		mockWorkers := new(mockWorkerPool)

		gf := &GoFlow{
			ctx:             ctx,
			cancel:          cancel,
			taskBroker:      taskBroker,
			resultsBroker:   resultBroker,
			results:         resultStore,
			workers:         mockWorkers,
			taskHandlers:    taskHandlers,
			resultsWriterWG: resultsWriterWG,
			started:         false,
		}

		mockWorkers.On("Start", ctx, taskBroker, resultBroker, taskHandlers).Once()

		returnCh := make(chan task.Result)
		expectedResult := task.Result{TaskID: "1234"}

		resultBroker.On("Dequeue", ctx).Twice().Return(channel.NewReadOnly(returnCh))

		resultStore.On("Put", expectedResult.TaskID, expectedResult).Once()

		// Act
		err := gf.Start()
		returnCh <- expectedResult

		cancel()

		// Assert
		resultsWriterWG.Wait()
		assert.Nil(t, err)
		mockWorkers.AssertExpectations(t)
		resultBroker.AssertExpectations(t)
		resultStore.AssertExpectations(t)
		assert.True(t, gf.started)
	})
}

func Test_GoFlow_RegisterHandler(t *testing.T) {
	t.Run("Puts the handler in the handler store if in local mode", func(t *testing.T) {
		// Arrange
		mockHandlers := new(mockKVStore[string, task.Handler])
		handler := func(_ any) task.Result {
			return task.Result{}
		}
		gf := GoFlow{
			taskHandlers: mockHandlers,
		}
		taskType := "exampleTask"

		mockHandlers.On("Put", taskType, mock.AnythingOfType("task.Handler")).Once()

		// Act
		gf.RegisterHandler(taskType, handler)

		// Assert
		mockHandlers.AssertExpectations(t)
	})

	t.Run("Doesn't put the handler in the handler store if in distributed mode", func(t *testing.T) {
		// Arrange
		mockHandlers := new(mockKVStore[string, task.Handler])
		gf := GoFlow{
			taskHandlers: nil,
		}
		taskType := "exampleTask"

		// Act
		gf.RegisterHandler(taskType, nil)

		// Assert
		mockHandlers.AssertNotCalled(t, "")
	})
}

func Test_GoFlow_Push(t *testing.T) {
	t.Run("Submits the task to the broker", func(t *testing.T) {
		// Arrange
		mockBroker := new(mockBroker[task.Task])

		ctx := context.Background()

		gf := GoFlow{
			ctx:        ctx,
			taskBroker: mockBroker,
			started:    true,
		}

		var submittedTask task.Task

		mockBroker.On("Submit", mock.Anything, mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
			submittedTask, _ = args.Get(1).(task.Task)
		})

		taskType := "exampleTask"
		payload := "examplePayload"

		// Act
		taskID, err := gf.Push(taskType, payload)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, submittedTask.ID, taskID)
		assert.Equal(t, taskType, submittedTask.Type)
		assert.Equal(t, payload, submittedTask.Payload)

		mockBroker.AssertExpectations(t)
	})

	t.Run("Returns an error if task submission fails", func(t *testing.T) {
		// Arrange
		mockBroker := new(mockBroker[task.Task])

		ctx := context.Background()

		gf := GoFlow{
			ctx:        ctx,
			taskBroker: mockBroker,
			started:    true,
		}

		submissionError := errors.New("submission error")
		mockBroker.On("Submit", mock.Anything, mock.Anything).Once().Return(submissionError)

		// Act
		_, err := gf.Push("exampleTask", "examplePayload")

		// Assert
		assert.EqualError(t, err, submissionError.Error())

		mockBroker.AssertExpectations(t)
	})

	t.Run("Returns ErrNotStarted if GoFlow instance is not started", func(t *testing.T) {
		// Arrange
		mockBroker := new(mockBroker[task.Task])

		ctx := context.Background()

		gf := GoFlow{
			ctx:        ctx,
			taskBroker: mockBroker,
			started:    false,
		}

		// Act
		_, err := gf.Push("exampleTask", "examplePayload")

		// Assert
		assert.EqualError(t, err, ErrNotStarted.Error())

		mockBroker.AssertExpectations(t)
	})
}

func Test_GoFlow_GetResult(t *testing.T) {
	t.Run("Returns the result of given taskID if it exists", func(t *testing.T) {
		// Arrange
		mockResults := new(mockKVStore[string, task.Result])

		gf := GoFlow{
			results: mockResults,
			started: true,
		}

		taskID := "taskID"

		expectedResult := task.Result{Payload: "result"}

		mockResults.On("Get", mock.Anything).Once().Return(expectedResult, true)

		// Act
		result, ok, err := gf.GetResult(taskID)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, expectedResult, result)
		assert.True(t, ok)
	})

	t.Run("Returns false if given taskID doesn't exist", func(t *testing.T) {
		// Arrange
		mockResults := new(mockKVStore[string, task.Result])

		gf := GoFlow{
			results: mockResults,
			started: true,
		}

		taskID := "taskID"

		expectedResult := task.Result{}

		mockResults.On("Get", mock.Anything).Once().Return(expectedResult, false)

		// Act
		result, ok, err := gf.GetResult(taskID)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, expectedResult, result)
		assert.False(t, ok)
	})

	t.Run("Returns ErrNotStarted if GoFlow instance isn't started", func(t *testing.T) {
		// Arrange
		mockResults := new(mockKVStore[string, task.Result])

		gf := GoFlow{
			results: mockResults,
			started: false,
		}

		taskID := "taskID"

		// Act
		result, ok, err := gf.GetResult(taskID)

		// Assert
		assert.EqualError(t, err, ErrNotStarted.Error())
		assert.False(t, ok)
		assert.Equal(t, task.Result{}, result)
		mockResults.AssertExpectations(t)
	})
}

func Test_GoFlow_Stop(t *testing.T) {
	t.Run("Calls cancel and waits for all components to shut down", func(t *testing.T) {
		// Arrange
		wasCancelCalled := false
		mockCancel := func() {
			wasCancelCalled = true
		}

		mockWorkerPool := &mockWorkerPool{}
		mockWorkerPool.On("AwaitShutdown").Once()

		mockTaskBroker := &mockBroker[task.Task]{}
		mockTaskBroker.On("AwaitShutdown").Once()

		mockResultBroker := &mockBroker[task.Result]{}
		mockResultBroker.On("AwaitShutdown").Once()

		gf := GoFlow{
			cancel:          mockCancel,
			workers:         mockWorkerPool,
			taskBroker:      mockTaskBroker,
			resultsBroker:   mockResultBroker,
			resultsWriterWG: &sync.WaitGroup{},
			started:         true,
		}

		// Act
		err := gf.Close()

		// Assert
		assert.Nil(t, err)
		assert.True(t, wasCancelCalled)

		mockWorkerPool.AssertExpectations(t)
		mockTaskBroker.AssertExpectations(t)
		mockResultBroker.AssertExpectations(t)
	})

	t.Run("Returns ErrNotStarted if GoFlow instance not started", func(t *testing.T) {
		// Arrange
		wasCancelCalled := false
		mockCancel := func() {
			wasCancelCalled = true
		}

		mockWorkerPool := &mockWorkerPool{}
		mockTaskBroker := &mockBroker[task.Task]{}
		mockResultBroker := &mockBroker[task.Result]{}

		gf := GoFlow{
			cancel:          mockCancel,
			workers:         mockWorkerPool,
			taskBroker:      mockTaskBroker,
			resultsBroker:   mockResultBroker,
			resultsWriterWG: &sync.WaitGroup{},
			started:         false,
		}

		// Act
		err := gf.Close()

		// Assert
		assert.EqualError(t, err, ErrNotStarted.Error())
		assert.False(t, wasCancelCalled)
		assert.False(t, gf.started)

		mockWorkerPool.AssertExpectations(t)
		mockTaskBroker.AssertExpectations(t)
		mockResultBroker.AssertExpectations(t)
	})
}

type mockWorkerPool struct {
	mock.Mock
}

func (m *mockWorkerPool) Start(
	ctx context.Context,
	taskQueue task.Dequeuer[task.Task],
	results task.Submitter[task.Result],
	taskHandlers workerpool.HandlerGetter,
) {
	m.Called(ctx, taskQueue, results, taskHandlers)
}

func (m *mockWorkerPool) AwaitShutdown() {
	m.Called()
}

type mockBroker[T any] struct {
	mock.Mock
}

func (m *mockBroker[T]) Submit(ctx context.Context, tsk T) error {
	args := m.Called(ctx, tsk)
	return args.Error(0)
}

func (m *mockBroker[T]) Dequeue(ctx context.Context) <-chan T {
	args := m.Called(ctx)
	return args.Get(0).(<-chan T)
}

func (m *mockBroker[T]) AwaitShutdown() {
	m.Called()
}

type mockKVStore[K comparable, V any] struct {
	mock.Mock
}

func (m *mockKVStore[K, V]) Put(key K, value V) {
	m.Called(key, value)
}

func (m *mockKVStore[K, V]) Get(key K) (V, bool) {
	args := m.Called(key)
	return args.Get(0).(V), args.Bool(1)
}
