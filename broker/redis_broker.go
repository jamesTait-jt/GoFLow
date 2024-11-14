package broker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jamesTait-jt/goflow/task"
	"github.com/redis/go-redis/v9"
)

type redisClient interface {
	LPush(ctx context.Context, key string, values ...any) *redis.IntCmd
	BRPop(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd
}

// Encoder defines methods for serializing and deserializing tasks of type T, where
// T satisfies task.TaskOrResult. This allows tasks and results to be encoded for
// sending over Redis and decoded when retrieved.
type Encoder[T task.TaskOrResult] interface {
	Serialise(toSerialise T) ([]byte, error)
	Deserialise(toDeserialise []byte) (T, error)
}

// RedisBroker is a Redis-backed message broker that supports submitting and
// asynchronously retrieving tasks of type T. It provides a channel-based interface for
// consuming tasks and includes options for configuring logging.
type RedisBroker[T task.TaskOrResult] struct {
	client        redisClient
	redisQueueKey string
	outChan       chan T
	started       sync.Once
	wg            *sync.WaitGroup
	encoder       Encoder[T]
	opts          redisBrokerOptions
}

// NewRedisBroker creates a new RedisBroker instance with the specified Redis client, queue key,
// and Encoder.
func NewRedisBroker[T task.TaskOrResult](
	client redisClient,
	key string,
	encoder Encoder[T],
	opt ...RedisBrokerOption,
) *RedisBroker[T] {
	opts := defaultRedisBrokerOptions()

	for _, o := range opt {
		o.apply(&opts)
	}

	return &RedisBroker[T]{
		client:        client,
		redisQueueKey: key,
		encoder:       encoder,
		opts:          opts,
		outChan:       make(chan T),
		wg:            &sync.WaitGroup{},
	}
}

// Submit serializes a task and pushes it to the Redis queue. If serialization or
// pushing fails, Submit returns an error.
func (rb *RedisBroker[T]) Submit(ctx context.Context, submission T) error {
	serialised, err := rb.encoder.Serialise(submission)
	if err != nil {
		return err
	}

	_, err = rb.client.LPush(ctx, rb.redisQueueKey, serialised).Result()
	if err != nil {
		return err
	}

	return nil
}

// Dequeue returns a receive-only channel that emits tasks as they are retrieved from
// the Redis queue. Dequeue starts a background goroutine for polling tasks from the
// queue. The polling process stops when the provided context is canceled. errors are
// logged but they will not stop the polling loop. pollRedis exits when the provided
// context is cancelled.
func (rb *RedisBroker[T]) Dequeue(ctx context.Context) <-chan T {
	rb.started.Do(func() {
		rb.wg.Add(1)
		go rb.pollRedis(ctx)
	})

	return rb.outChan
}

func (rb *RedisBroker[T]) pollRedis(ctx context.Context) {
	defer rb.wg.Done()

	for {
		redisResult, err := rb.client.BRPop(ctx, 0, rb.redisQueueKey).Result()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			rb.opts.logger.Warn(err.Error())

			continue
		}

		result, err := rb.encoder.Deserialise([]byte(redisResult[1]))
		if err != nil {
			rb.opts.logger.Warn(err.Error())

			continue
		}

		rb.outChan <- result
	}
}

// AwaitShutdown waits for the background polling goroutine to finish.
// This method should be called during shutdown to ensure all resources are released.
func (rb *RedisBroker[T]) AwaitShutdown() {
	rb.wg.Wait()
}
