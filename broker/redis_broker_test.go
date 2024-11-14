//go:build unit

package broker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jamesTait-jt/goflow/pkg/channel"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_NewRedisBroker(t *testing.T) {
	t.Run("Creates a redis broker with the variables initialised", func(t *testing.T) {
		// Arrange
		client := new(mockRedisClient)
		key := "queue"
		serialiser := new(mockSerialiser[task.Task])
		deserialiser := new(mockDeserialiser[task.Task])
		logger := new(log.TestifyMock)

		// Act
		broker := NewRedisBroker(
			client,
			key,
			serialiser,
			deserialiser,
			WithLogger(logger),
			WithPollTimeout(time.Second),
		)

		// Assert
		assert.Equal(t, client, broker.client)
		assert.Equal(t, key, broker.redisQueueKey)
		assert.Equal(t, serialiser, broker.serialiser)
		assert.Equal(t, deserialiser, broker.deserialiser)
		assert.NotNil(t, broker.outChan)
		assert.NotNil(t, &broker.started)
		assert.NotNil(t, broker.wg)
		assert.Equal(t, time.Second, broker.opts.pollTimeout)
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

		b := NewRedisBroker(
			mockClient,
			queueKey,
			serialiser,
			nil,
		)

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

		b := NewRedisBroker(
			mockClient,
			queueKey,
			serialiser,
			nil,
		)

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

		b := NewRedisBroker(
			mockClient,
			queueKey,
			serialiser,
			nil,
		)

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

func Test_RedisBroker_Dequeue(t *testing.T) {
	t.Run("Polls redis and returns a channel for listening", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())

		mockClient := new(mockRedisClient)
		queueKey := "queue"
		deserialiser := new(mockDeserialiser[task.Task])
		pollTimeout := time.Millisecond

		br := NewRedisBroker(
			mockClient,
			queueKey,
			nil,
			deserialiser,
			WithPollTimeout(pollTimeout),
		)

		returnedFromRedis := &redis.StringSliceCmd{}
		returnedFromRedis.SetVal([]string{"", "returned val"})
		mockClient.On("BRPop", ctx, pollTimeout, []string{queueKey}).Once().Return(returnedFromRedis)

		deserialisedVal := task.Task{}
		deserialiser.On("Deserialise", []byte(returnedFromRedis.Val()[1])).Once().Run(func(args mock.Arguments) {
			// Cancel so that there will only be one iteration of the polling loop
			cancel()
		}).Return(deserialisedVal, nil)

		// Act
		c := br.Dequeue(ctx)

		// Assert
		assert.Equal(t, c, channel.NewReadOnly(br.outChan))
		assert.Equal(t, deserialisedVal, <-c)

		br.wg.Wait()
		mockClient.AssertExpectations(t)
	})

	t.Run("Handles Redis timeout (redis.Nil) and continues polling", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())

		mockClient := new(mockRedisClient)
		queueKey := "queue"
		deserialiser := new(mockDeserialiser[task.Task])
		pollTimeout := time.Millisecond

		br := NewRedisBroker(
			mockClient,
			queueKey,
			nil,
			deserialiser,
			WithPollTimeout(pollTimeout),
		)

		returnedResult := &redis.StringSliceCmd{}
		returnedResult.SetErr(redis.Nil)
		mockClient.On("BRPop", ctx, pollTimeout, []string{queueKey}).Once().Run(func(args mock.Arguments) {
			// Cancel so that there will only be one iteration of the polling loop
			cancel()
		}).Return(returnedResult)

		// Act
		br.Dequeue(ctx)

		// Assert
		br.wg.Wait()

		select {
		case <-br.outChan:
			t.Error("Expected no value in outChan")
		default:
			// Test passes, as no task should be available
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("Logs Redis error and continues polling", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())

		mockClient := new(mockRedisClient)
		queueKey := "queue"
		deserialiser := new(mockDeserialiser[task.Task])
		pollTimeout := time.Millisecond
		logger := new(log.TestifyMock)

		br := NewRedisBroker(
			mockClient,
			queueKey,
			nil,
			deserialiser,
			WithPollTimeout(pollTimeout),
			WithLogger(logger),
		)

		returnedResult := &redis.StringSliceCmd{}
		returnedResult.SetErr(redis.ErrClosed)
		mockClient.On("BRPop", ctx, pollTimeout, []string{queueKey}).Once().Run(func(args mock.Arguments) {
			// Cancel so that there will only be one iteration of the polling loop
			cancel()
		}).Return(returnedResult)

		logger.On("Warn", "BRPop error: redis: client is closed").Once()

		// Act
		br.Dequeue(ctx)

		// Assert
		br.wg.Wait()

		select {
		case <-br.outChan:
			t.Error("Expected no value in outChan")
		default:
			// Test passes, as no task should be available
		}

		mockClient.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("Handles deserialisation failure", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())

		mockClient := new(mockRedisClient)
		queueKey := "queue"
		deserialiser := new(mockDeserialiser[task.Task])
		pollTimeout := time.Millisecond
		logger := new(log.TestifyMock)

		br := NewRedisBroker(
			mockClient,
			queueKey,
			nil,
			deserialiser,
			WithPollTimeout(pollTimeout),
			WithLogger(logger),
		)

		returnedFromRedis := &redis.StringSliceCmd{}
		returnedFromRedis.SetVal([]string{"", "faulty data"})
		mockClient.On("BRPop", ctx, pollTimeout, []string{queueKey}).Once().Return(returnedFromRedis)

		deserialiser.On("Deserialise", []byte("faulty data")).Once().Run(func(args mock.Arguments) {
			// Cancel so that there will only be one iteration of the polling loop
			cancel()
		}).Return(task.Task{}, fmt.Errorf("deserialisation error"))

		logger.On("Warn", "Failed to deserialize task: deserialisation error").Once()

		// Act
		br.Dequeue(ctx)

		// Assert
		br.wg.Wait()

		select {
		case <-br.outChan:
			t.Error("Expected no value in outChan due to deserialization failure")
		default:
			// Test passes, as no task should be available
		}

		mockClient.AssertExpectations(t)
		deserialiser.AssertExpectations(t)
		logger.AssertExpectations(t)
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
