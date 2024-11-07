package service

import (
	"context"
	"testing"

	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/jamesTait-jt/goflow/workerpool"
	"github.com/stretchr/testify/mock"
)

func Test_WorkerpoolService_Start(t *testing.T) {
	ctx := context.Background()

	// Arrange
	pool := new(mockWorkerpoolRunner)
	taskQueue := broker.NewChannelBroker[task.Task](0)
	resultQueue := broker.NewChannelBroker[task.Result](0)
	taskHandlers := store.NewInMemoryKVStore[string, task.Handler]()

	service := NewWorkerpoolService(pool, taskQueue, resultQueue, taskHandlers)

	pool.On("Start", ctx, taskQueue, resultQueue, taskHandlers).Once()
	pool.On("AwaitShutdown").Once()

	// Act
	service.Start(ctx)

	// Assert
	pool.AssertExpectations(t)
}

type mockWorkerpoolRunner struct {
	mock.Mock
}

func (m *mockWorkerpoolRunner) Start(ctx context.Context, taskQueue task.Dequeuer[task.Task], results task.Submitter[task.Result], taskHandlers workerpool.HandlerGetter) {
	m.Called(ctx, taskQueue, results, taskHandlers)
}

func (m *mockWorkerpoolRunner) AwaitShutdown() {
	m.Called()
}
