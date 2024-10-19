package goflow

import (
	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/pkg/task"
)

type Option interface {
	apply(*options)
}

type options struct {
	taskHandlerStore KVStore[string, task.Handler]
	resultsStore     KVStore[string, task.Result]
	taskBroker       Broker[task.Task]
	resultBroker     Broker[task.Result]
	workerPool       WorkerPool
}

func defaultOptions() options {
	defaultTaskBrokerChannelSize := 10

	return options{
		resultsStore: store.NewInMemoryKVStore[string, task.Result](),
		taskBroker:   broker.NewChannelBroker[task.Task](defaultTaskBrokerChannelSize),
		resultBroker: broker.NewChannelBroker[task.Result](0),
	}
}

type resultsStoreOption struct {
	ResultsStore KVStore[string, task.Result]
}

func (r resultsStoreOption) apply(opts *options) {
	opts.resultsStore = r.ResultsStore
}

func WithResultsStore(resultsStore KVStore[string, task.Result]) Option {
	return resultsStoreOption{ResultsStore: resultsStore}
}

type taskBrokerOption struct {
	TaskBroker Broker[task.Task]
}

func (t taskBrokerOption) apply(opts *options) {
	opts.taskBroker = t.TaskBroker
}

func WithTaskBroker(taskBroker Broker[task.Task]) Option {
	return taskBrokerOption{TaskBroker: taskBroker}
}

type resultBrokerOption struct {
	ResultBroker Broker[task.Result]
}

func (r resultBrokerOption) apply(opts *options) {
	opts.resultBroker = r.ResultBroker
}

func WithResultBroker(taskBroker Broker[task.Result]) Option {
	return resultBrokerOption{ResultBroker: taskBroker}
}

type localModeOption struct {
	WorkerPool       WorkerPool
	TaskHandlerStore KVStore[string, task.Handler]
}

func (l localModeOption) apply(opts *options) {
	opts.workerPool = l.WorkerPool
	opts.taskHandlerStore = l.TaskHandlerStore
}

func WithLocalMode(workerPool WorkerPool, taskHandlerStore KVStore[string, task.Handler]) Option {
	return localModeOption{
		WorkerPool:       workerPool,
		TaskHandlerStore: taskHandlerStore,
	}
}
