package goflow

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/jamesTait-jt/goflow/workerpool"
)

// Broker is an interface that abstracts messaging systems used by GoFlow.
// It requires two brokers: one for submitting tasks to the worker pool,
// and another for receiving results from the worker pool.
type Broker[T task.TaskOrResult] interface {
	task.Submitter[T]
	task.Dequeuer[T]
	AwaitShutdown()
}

// WorkerPool is implemented only when running GoFlow in local mode. In distributed
// mode, the worker pool is abstracted away from GoFlow by the task and results brokers.
type WorkerPool interface {
	// Start initializes the worker pool, with workers listening to taskQueue and
	// submitting results. It should be non-blocking, starting workers in their own
	// goroutines and returning immediately. The worker pool will run until the context
	// is canceled.
	Start(
		ctx context.Context,
		taskQueue task.Dequeuer[task.Task],
		results task.Submitter[task.Result],
		taskHandlers workerpool.HandlerGetter,
	)

	// AwaitShutdown ensures that all workers complete processing after GoFlow's context
	// is canceled, allowing for graceful shutdown without leaving hanging goroutines.
	AwaitShutdown()
}

// KVStore defines a key-value store interface in the GoFlow framework. It provides
// methods for storing and retrieving values associated with keys.
//
// Users can implement KVStore to create custom key-value storage solutions as needed.
// Example implementations could include in-memory, database-backed, or other forms
// of key-value mappings.
type KVStore[K comparable, V any] interface {
	// Put stores the value associated with the given key.
	Put(k K, v V)

	// Get retrieves the value associated with the given key, returning
	// the value and a boolean indicating whether the key was found.
	Get(k K) (V, bool)
}

// GoFlow is the core structure of the framework. It manages interactions with brokers
// to send tasks and receive results. GoFlow continually polls the results broker,
// writing incoming results to the results store.
//
// In local mode, GoFlow also manages the worker pool and task handler registry.
type GoFlow struct {
	ctx             context.Context
	cancel          context.CancelFunc
	workers         WorkerPool
	taskBroker      Broker[task.Task]
	taskHandlers    KVStore[string, task.Handler]
	resultsBroker   Broker[task.Result]
	results         KVStore[string, task.Result]
	resultsWriterWG *sync.WaitGroup
	started         bool
}

var (
	ErrAlreadyStarted = errors.New("GoFlow is already started")
	ErrNotStarted     = errors.New("GoFlow is not started yet")
)

// New creates and initializes a new GoFlow instance in distributed mode.
// It sets up the context for cancellation and configures the necessary components.
// If no options are provided, default values are used (see defaultOptions()).
//
// For detailed configuration options, see options.go.
func New(taskBroker Broker[task.Task], resultsBroker Broker[task.Result], opts ...Option) *GoFlow {
	options := defaultOptions()

	for _, o := range opts {
		o.apply(&options)
	}

	ctx, cancel := context.WithCancel(context.Background())

	gf := GoFlow{
		ctx:             ctx,
		cancel:          cancel,
		taskBroker:      taskBroker,
		resultsBroker:   resultsBroker,
		results:         options.resultsStore,
		resultsWriterWG: &sync.WaitGroup{},
	}

	return &gf
}

// NewLocalMode creates and initializes a new GoFlow instance configured for local mode.
// It sets up a worker pool and task/result brokers with specified sizes for task and
// result queues. The context is also set up for cancellation, and if no options are
// provided, default values are used (see defaultOptions()).
//
// For detailed configuration options, see options.go.
func NewLocalMode(
	numWorkers, taskQueueSize, resultQueueSize int,
	taskHandlers KVStore[string, task.Handler],
	opts ...Option,
) *GoFlow {
	options := defaultOptions()

	for _, o := range opts {
		o.apply(&options)
	}

	ctx, cancel := context.WithCancel(context.Background())

	gf := GoFlow{
		ctx:             ctx,
		cancel:          cancel,
		workers:         workerpool.New(numWorkers),
		taskBroker:      broker.NewChannelBroker[task.Task](taskQueueSize),
		taskHandlers:    taskHandlers,
		resultsBroker:   broker.NewChannelBroker[task.Result](resultQueueSize),
		results:         options.resultsStore,
		resultsWriterWG: &sync.WaitGroup{},
	}

	return &gf
}

// Start initiates the execution of the GoFlow instance. It checks if the worker pool
// and task handlers are configured (i.e., running in local mode). If so, it starts
// the worker pool to process tasks submitted through the task broker, which is not
// necessary in distributed mode.
//
// Additionally, the method launches a goroutine to persist results from the results
// broker to the results store.
func (gf *GoFlow) Start() error {
	if gf.started {
		return ErrAlreadyStarted
	}

	gf.started = true

	// Running with local worker pool
	if gf.workers != nil && gf.taskHandlers != nil {
		gf.workers.Start(gf.ctx, gf.taskBroker, gf.resultsBroker, gf.taskHandlers)
	}

	gf.resultsWriterWG.Add(1)
	go gf.persistResults(gf.resultsBroker, gf.resultsWriterWG)

	return nil
}

// Close gracefully shuts down the GoFlow instance. It cancels the context to signal
// all ongoing operations to stop. If the worker pool is configured, (i.e. local mode)
// it waits for all workers to complete their tasks and shut down before returning.
func (gf *GoFlow) Close() error {
	if !gf.started {
		return ErrNotStarted
	}

	gf.started = false

	gf.cancel()

	gf.resultsWriterWG.Wait()
	gf.resultsBroker.AwaitShutdown()
	gf.taskBroker.AwaitShutdown()

	if gf.workers != nil {
		gf.workers.AwaitShutdown()
	}

	return nil
}

// RegisterHandler registers a task handler for the specified task type. It stores
// the handler in the taskHandlers store for local mode execution.
//
// If the GoFlow instance is not running in local mode (i.e., taskHandlers is nil),
// a warning is logged, and the handler is not registered. In distributed mode,
// handlers must be pre-registered when compiling the worker pool.
func (gf *GoFlow) RegisterHandler(taskType string, handler task.Handler) {
	if gf.taskHandlers == nil {
		log.Println("handlers can only be registered in local mode")

		return
	}

	gf.taskHandlers.Put(taskType, handler)
}

// Push submits a new task with the specified type and payload to the task broker.
// It creates a task, submits it to the broker, and returns the task's ID.
//
// The task is processed by the worker pool, and the caller can use the returned
// task ID to retrieve the result later.
func (gf *GoFlow) Push(taskType string, payload any) (string, error) {
	if !gf.started {
		return "", ErrNotStarted
	}

	t := task.New(taskType, payload)

	err := gf.taskBroker.Submit(gf.ctx, t)
	if err != nil {
		return "", err
	}

	return t.ID, nil
}

// GetResult retrieves the result associated with the specified task ID. It returns
// the result and a boolean indicating whether the result was found.
//
// If the task with the given ID has completed, the result will be returned. If the
// task has not yet completed or does not exist, the boolean will be false.
func (gf *GoFlow) GetResult(taskID string) (task.Result, bool, error) {
	if !gf.started {
		return task.Result{}, false, ErrNotStarted
	}

	result, ok := gf.results.Get(taskID)

	return result, ok, nil
}

func (gf *GoFlow) persistResults(results task.Dequeuer[task.Result], wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-gf.ctx.Done():
			return

		case result := <-results.Dequeue(gf.ctx):
			gf.results.Put(result.TaskID, result)
		}
	}
}
