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
		encoder := new(mockEncoder[task.Task])
		logger := new(log.TestifyMock)

		// Act
		broker := NewRedisBroker(
			client,
			key,
			encoder,
			WithLogger(logger),
			WithPollTimeout(time.Second),
		)

		// Assert
		assert.Equal(t, client, broker.client)
		assert.Equal(t, key, broker.redisQueueKey)
		assert.Equal(t, encoder, broker.encoder)
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
		encoder := new(mockEncoder[task.Task])

		b := NewRedisBroker(mockClient, queueKey, encoder)

		serialised := []byte{1, 2, 3, 4}
		encoder.On("Serialise", tsk).Return(serialised, nil)

		returnedCmd := &redis.IntCmd{}
		returnedCmd.SetErr(nil)
		mockClient.On("LPush", ctx, queueKey, []interface{}{serialised}).Return(returnedCmd)

		// Act
		err := b.Submit(ctx, tsk)

		// Assert
		assert.NoError(t, err)
		encoder.AssertExpectations(t)
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
		encoder := new(mockEncoder[task.Task])

		b := NewRedisBroker(mockClient, queueKey, encoder)

		serialiserError := errors.New("failed to serialise")
		encoder.On("Serialise", tsk).Return([]byte{}, serialiserError)

		// Act
		err := b.Submit(ctx, tsk)

		// Assert
		assert.EqualError(t, err, serialiserError.Error())
		encoder.AssertExpectations(t)
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
		encoder := new(mockEncoder[task.Task])

		b := NewRedisBroker(mockClient, queueKey, encoder)

		serialised := []byte{1, 2, 3, 4}
		encoder.On("Serialise", tsk).Return(serialised, nil)

		lpushErr := errors.New("lpush error")
		returnedCmd := &redis.IntCmd{}
		returnedCmd.SetErr(lpushErr)
		mockClient.On("LPush", ctx, queueKey, []interface{}{serialised}).Return(returnedCmd)

		// Act
		err := b.Submit(ctx, tsk)

		// Assert
		assert.EqualError(t, err, lpushErr.Error())
		encoder.AssertExpectations(t)
	})
}

func Test_RedisBroker_Dequeue(t *testing.T) {
	t.Run("Polls redis and returns a channel for listening", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		mockClient := new(mockRedisClient)
		queueKey := "queue"
		encoder := new(mockEncoder[task.Task])
		pollTimeout := time.Millisecond

		br := NewRedisBroker(mockClient, queueKey, encoder, WithPollTimeout(pollTimeout))

		returnedFromRedis := &redis.StringSliceCmd{}
		returnedFromRedis.SetVal([]string{"", "returned val"})
		mockClient.On("BRPop", ctx, time.Duration(0), []string{queueKey}).Once().Return(returnedFromRedis)

		deserialisedVal := task.Task{}
		encoder.On("Deserialise", []byte(returnedFromRedis.Val()[1])).Once().Return(deserialisedVal, nil)

		errReturnedFromRedis := &redis.StringSliceCmd{}
		errReturnedFromRedis.SetErr(context.Canceled)
		mockClient.On("BRPop", ctx, time.Duration(0), []string{queueKey}).Once().Return(errReturnedFromRedis)

		// Act
		c := br.Dequeue(ctx)

		// Assert
		assert.Equal(t, c, channel.NewReadOnly(br.outChan))
		assert.Equal(t, deserialisedVal, <-c)

		br.wg.Wait()
		mockClient.AssertExpectations(t)
	})

	t.Run("Logs Redis error and continues polling", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		mockClient := new(mockRedisClient)
		queueKey := "queue"
		encoder := new(mockEncoder[task.Task])
		pollTimeout := time.Millisecond
		logger := new(log.TestifyMock)

		br := NewRedisBroker(mockClient, queueKey, encoder, WithPollTimeout(pollTimeout), WithLogger(logger))

		returnedResult := &redis.StringSliceCmd{}
		returnedResult.SetErr(redis.ErrClosed)
		mockClient.On("BRPop", ctx, time.Duration(0), []string{queueKey}).Once().Return(returnedResult)

		logger.On("Warn", "redis: client is closed").Once()

		errReturnedFromRedis := &redis.StringSliceCmd{}
		errReturnedFromRedis.SetErr(context.Canceled)
		mockClient.On("BRPop", ctx, time.Duration(0), []string{queueKey}).Once().Return(errReturnedFromRedis)

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
		encoder := new(mockEncoder[task.Task])
		pollTimeout := time.Millisecond
		logger := new(log.TestifyMock)

		br := NewRedisBroker(mockClient, queueKey, encoder, WithPollTimeout(pollTimeout), WithLogger(logger))

		returnedFromRedis := &redis.StringSliceCmd{}
		returnedFromRedis.SetVal([]string{"", "faulty data"})
		mockClient.On("BRPop", ctx, time.Duration(0), []string{queueKey}).Once().Return(returnedFromRedis)

		encoder.On("Deserialise", []byte("faulty data")).Once().Run(func(args mock.Arguments) {
			// Cancel so that there will only be one iteration of the polling loop
			cancel()
		}).Return(task.Task{}, fmt.Errorf("deserialisation error"))

		logger.On("Warn", "deserialisation error").Once()

		errReturnedFromRedis := &redis.StringSliceCmd{}
		errReturnedFromRedis.SetErr(context.Canceled)
		mockClient.On("BRPop", ctx, time.Duration(0), []string{queueKey}).Once().Return(errReturnedFromRedis)

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
		encoder.AssertExpectations(t)
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

type mockEncoder[T task.TaskOrResult] struct {
	mock.Mock
}

func (m *mockEncoder[T]) Serialise(toSerialise T) ([]byte, error) {
	args := m.Called(toSerialise)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockEncoder[T]) Deserialise(toDeserialise []byte) (T, error) {
	args := m.Called(toDeserialise)
	return args.Get(0).(T), args.Error(1)
}
