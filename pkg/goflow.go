package goflow

import (
	"context"
	"log"

	"github.com/jamesTait-jt/goflow/pkg/task"
	"github.com/jamesTait-jt/goflow/pkg/workerpool"
)

// Broker is an interface that abstracts messaging systems used by GoFlow.
// It requires two brokers: one for submitting tasks to the worker pool,
// and another for receiving results from the worker pool.
type Broker[T task.TaskOrResult] interface {
	task.Submitter[T]
	task.Dequeuer[T]
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
	ctx           context.Context
	cancel        context.CancelFunc
	workers       WorkerPool
	taskBroker    Broker[task.Task]
	taskHandlers  KVStore[string, task.Handler]
	resultsBroker Broker[task.Result]
	results       KVStore[string, task.Result]
}

// New creates and initializes a new GoFlow instance with the provided options.
// It sets up the context for cancellation and configures the necessary components.
// If no options are provided, default values are used (see defaultOptions()).
//
// The following configuration options can be specified using the corresponding
// functions:
//   - WithResultsStore: Sets a custom results store for task results.
//   - WithTaskBroker: Sets a custom broker for submitting tasks.
//   - WithResultBroker: Sets a custom broker for receiving results.
//   - WithLocalMode: Configures the worker pool and task handler store for local mode.
func New(opts ...Option) *GoFlow {
	options := defaultOptions()

	for _, o := range opts {
		o.apply(&options)
	}

	ctx, cancel := context.WithCancel(context.Background())

	gf := GoFlow{
		ctx:           ctx,
		cancel:        cancel,
		workers:       options.workerPool,
		taskBroker:    options.taskBroker,
		taskHandlers:  options.taskHandlerStore,
		results:       options.resultsStore,
		resultsBroker: options.resultBroker,
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
func (gf *GoFlow) Start() {
	// Running with local worker pool
	if gf.workers != nil && gf.taskHandlers != nil {
		gf.workers.Start(gf.ctx, gf.taskBroker, gf.resultsBroker, gf.taskHandlers)
	}

	go gf.persistResults(gf.resultsBroker)
}

// Stop gracefully shuts down the GoFlow instance. It cancels the context to signal
// all ongoing operations to stop. If the worker pool is configured, (i.e. local mode)
// it waits for all workers to complete their tasks and shut down before returning.
func (gf *GoFlow) Stop() {
	gf.cancel()

	if gf.workers != nil {
		gf.workers.AwaitShutdown()
	}
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
	t := task.New(taskType, payload)

	gf.taskBroker.Submit(gf.ctx, t)

	return t.ID, nil
}

// GetResult retrieves the result associated with the specified task ID. It returns
// the result and a boolean indicating whether the result was found.
//
// If the task with the given ID has completed, the result will be returned. If the
// task has not yet completed or does not exist, the boolean will be false.
func (gf *GoFlow) GetResult(taskID string) (task.Result, bool) {
	result, ok := gf.results.Get(taskID)
	return result, ok
}

func (gf *GoFlow) persistResults(results task.Dequeuer[task.Result]) {
	for {
		select {
		case <-gf.ctx.Done():
			return

		case result := <-results.Dequeue(gf.ctx):
			gf.results.Put(result.TaskID, result)
		}
	}
}
