//go:build integration

package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/jamesTait-jt/goflow/cmd/workerpool/config"
	"github.com/jamesTait-jt/goflow/cmd/workerpool/runtime"
	"github.com/jamesTait-jt/goflow/pkg/serialise"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestWorkerpool_Integration(t *testing.T) {
	// Arrange
	ctx := context.Background()

	redisContainer, err := startRedisContainer(ctx)
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, redisContainer)

	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	client, err := connectToRedisContainer(ctx, endpoint)
	require.NoError(t, err)
	defer client.Close()

	taskSerialiser := serialise.NewGobSerialiser[task.Task]()
	resultSerialiser := serialise.NewGobSerialiser[task.Result]()

	relHandlersPath := "./testdata/handlers"
	absHandlersPath, err := filepath.Abs(relHandlersPath)
	require.NoError(t, err)

	workerpoolConfig := &config.Config{
		NumWorkers:   1,
		HandlersPath: absHandlersPath,
		BrokerType:   "redis",
		BrokerAddr:   endpoint,
	}
	workerpoolRuntime := runtime.Runtime{
		Conf: workerpoolConfig,
	}

	go func() {
		runtimeErr := workerpoolRuntime.Run()
		require.NoError(t, runtimeErr)
	}()

	t.Run("Reads from tasks queue and publishes result to results queue", func(t *testing.T) {
		// Arrange
		doubleMe := 10
		doubleTask := task.New("doubler", doubleMe)
		serialisedDouble, err := taskSerialiser.Serialise(doubleTask)
		require.NoError(t, err)

		repeatMe := "saka ballon d'or 2026"
		repeatTask := task.New("repeater", repeatMe)
		serialisedRepeat, err := taskSerialiser.Serialise(repeatTask)
		require.NoError(t, err)

		// Act
		_, err = client.LPush(ctx, "tasks", serialisedDouble).Result()
		require.NoError(t, err)

		redisResult, err := client.BRPop(ctx, 5*time.Second, "results").Result()
		require.NoError(t, err)

		doubleResult, err := resultSerialiser.Deserialise([]byte(redisResult[1]))
		require.NoError(t, err)

		_, err = client.LPush(ctx, "tasks", serialisedRepeat).Result()
		require.NoError(t, err)

		redisResult, err = client.BRPop(ctx, 5*time.Second, "results").Result()
		require.NoError(t, err)

		repeatResult, err := resultSerialiser.Deserialise([]byte(redisResult[1]))
		require.NoError(t, err)

		// Assert
		assert.Equal(t, 2*doubleMe, doubleResult.Payload)
		assert.Equal(t, doubleTask.ID, doubleResult.TaskID)

		assert.Equal(t, repeatMe, repeatResult.Payload)
		assert.Equal(t, repeatTask.ID, repeatResult.TaskID)
	})

}

func startRedisContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(30 * time.Second),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, err
	}

	return redisC, nil
}

func connectToRedisContainer(ctx context.Context, endpoint string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return client, nil
}
