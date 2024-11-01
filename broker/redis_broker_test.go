package broker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jamesTait-jt/goflow/task"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_NewRedisBroker(t *testing.T) {
	t.Run("Creates a redis broker with the variables initialised", func(t *testing.T) {
		// Arrange
		client := &mockRedisClient{}
		key := "queue"
		serialiser := &mockSerialiser[task.Task]{}
		deserialiser := &mockDeserialiser[task.Task]{}

		// Act
		broker := NewRedisBroker[task.Task](client, key, serialiser, deserialiser)

		// Assert
		assert.Equal(t, client, broker.client)
		assert.Equal(t, key, broker.redisQueueKey)
		assert.Equal(t, serialiser, broker.serialiser)
		assert.Equal(t, deserialiser, broker.deserialiser)
		assert.NotNil(t, broker.outChan)
		assert.NotNil(t, &broker.started)
	})
}

func Test_RedisBroker_Submit(t *testing.T) {
	t.Run("Serialises the task and places it on the task queue", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		tsk := task.Task{
			ID: "randomID",
		}

		mockClient := new(mockRedisClient)
		queueKey := "queue"
		serialiser := new(mockSerialiser[task.Task])

		b := RedisBroker[task.Task]{
			client:        mockClient,
			redisQueueKey: queueKey,
			serialiser:    serialiser,
		}

		serialised := []byte{1, 2, 3, 4}
		serialiser.On("Serialise", tsk).Return(serialised, nil)

		returnedCmd := &redis.IntCmd{}
		returnedCmd.SetErr(nil)
		mockClient.On("LPush", ctx, queueKey, []interface{}{serialised}).Return(returnedCmd)

		// Act
		err := b.Submit(ctx, tsk)

		// Assert
		assert.NoError(t, err)
		serialiser.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("Does not push to redis if serialisation fails and returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		tsk := task.Task{
			ID: "randomID",
		}

		mockClient := new(mockRedisClient)
		queueKey := "queue"
		serialiser := new(mockSerialiser[task.Task])

		b := RedisBroker[task.Task]{
			client:        mockClient,
			redisQueueKey: queueKey,
			serialiser:    serialiser,
		}

		serialiserError := errors.New("failed to serialise")
		serialiser.On("Serialise", tsk).Return([]byte{}, serialiserError)

		// Act
		err := b.Submit(ctx, tsk)

		// Assert
		assert.EqualError(t, err, serialiserError.Error())
		serialiser.AssertExpectations(t)
		mockClient.AssertNotCalled(t, "LPush")
	})

	t.Run("Returns error if push to redis fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		tsk := task.Task{
			ID: "randomID",
		}

		mockClient := new(mockRedisClient)
		queueKey := "queue"
		serialiser := new(mockSerialiser[task.Task])

		b := RedisBroker[task.Task]{
			client:        mockClient,
			redisQueueKey: queueKey,
			serialiser:    serialiser,
		}

		serialised := []byte{1, 2, 3, 4}
		serialiser.On("Serialise", tsk).Return(serialised, nil)

		lpushErr := errors.New("lpush error")
		returnedCmd := &redis.IntCmd{}
		returnedCmd.SetErr(lpushErr)
		mockClient.On("LPush", ctx, queueKey, []interface{}{serialised}).Return(returnedCmd)

		// Act
		err := b.Submit(ctx, tsk)

		// Assert
		assert.EqualError(t, err, lpushErr.Error())
		serialiser.AssertExpectations(t)
	})
}

type mockRedisClient struct {
	mock.Mock
}

func (m *mockRedisClient) LPush(ctx context.Context, key string, values ...any) *redis.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (m *mockRedisClient) BRPop(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	args := m.Called(ctx, timeout, keys)
	return args.Get(0).(*redis.StringSliceCmd)
}

type mockSerialiser[T task.TaskOrResult] struct {
	mock.Mock
}

func (m *mockSerialiser[T]) Serialise(toSerialise T) ([]byte, error) {
	args := m.Called(toSerialise)
	return args.Get(0).([]byte), args.Error(1)
}

type mockDeserialiser[T task.TaskOrResult] struct {
	mock.Mock
}

func (m *mockDeserialiser[T]) Deserialise(toDeserialise []byte) (T, error) {
	args := m.Called(toDeserialise)
	return args.Get(0).(T), args.Error(1)
}
