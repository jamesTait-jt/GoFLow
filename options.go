package goflow

import (
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
)

var (
	defaultNumWorkers            = 5
	defaultTaskQueueBufferSize   = 0
	defaultResultQueueBufferSize = 0
)

type Option interface {
	apply(*options)
}

type options struct {
	numWorkers            int
	taskQueueBufferSize   int
	resultQueueBufferSize int
	resultsStore          KVStore[string, task.Result]
}

func defaultOptions() options {
	return options{
		numWorkers:            defaultNumWorkers,
		taskQueueBufferSize:   defaultTaskQueueBufferSize,
		resultQueueBufferSize: defaultResultQueueBufferSize,
		resultsStore:          store.NewInMemoryKVStore[string, task.Result](),
	}
}

type numWorkersOption struct {
	NumWorkers int
}

func (n numWorkersOption) apply(opts *options) {
	opts.numWorkers = n.NumWorkers
}

// WithNumWorkers allows you to set the number of goroutines that will spawn and listen
// to the task queue. Has no effect if running in distributed mode.
func WithNumWorkers(numWorkers int) Option {
	return numWorkersOption{NumWorkers: numWorkers}
}

type taskQueueBufferSizeOption struct {
	TaskQueueBufferSize int
}

func (t taskQueueBufferSizeOption) apply(opts *options) {
	opts.taskQueueBufferSize = t.TaskQueueBufferSize
}

// WithTaskQueueBufferSize allows you to set the buffer size of the task queue channel.
// Has no effect if running in distributed mode.
func WithTaskQueueBufferSize(bufferSize int) Option {
	return taskQueueBufferSizeOption{TaskQueueBufferSize: bufferSize}
}

type resultQueueBufferSizeOption struct {
	ResultQueueBufferSize int
}

func (r resultQueueBufferSizeOption) apply(opts *options) {
	opts.resultQueueBufferSize = r.ResultQueueBufferSize
}

// WithTaskQueueBufferSize allows you to set the buffer size of the result queue channel.
// Has no effect if running in distributed mode.
func WithResultQueueBufferSize(bufferSize int) Option {
	return resultQueueBufferSizeOption{ResultQueueBufferSize: bufferSize}
}

type resultsStoreOption struct {
	ResultsStore KVStore[string, task.Result]
}

func (r resultsStoreOption) apply(opts *options) {
	opts.resultsStore = r.ResultsStore
}

// WithResultsStore allows you to inject your own results store. Anything that implements
// the KVStore interface is viable.
func WithResultsStore(resultsStore KVStore[string, task.Result]) Option {
	return resultsStoreOption{ResultsStore: resultsStore}
}
