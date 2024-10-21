package goflow

import (
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
)

type Option interface {
	apply(*options)
}

type options struct {
	resultsStore KVStore[string, task.Result]
}

func defaultOptions() options {
	return options{
		resultsStore: store.NewInMemoryKVStore[string, task.Result](),
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
