//go:build integration

package integration

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/jamesTait-jt/goflow/cmd/workerpool/config"
	"github.com/jamesTait-jt/goflow/cmd/workerpool/runtime"
	"github.com/jamesTait-jt/goflow/pkg/serialise"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
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
		doubleMe := `{"N":10}`
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
		assert.Equal(t, 20, doubleResult.Payload)
		assert.Equal(t, doubleTask.ID, doubleResult.TaskID)

		assert.Equal(t, repeatMe, repeatResult.Payload)
		assert.Equal(t, repeatTask.ID, repeatResult.TaskID)
	})

	t.Run("Handles plugins that return errors", func(t *testing.T) {
		// Arrange
		errorMe := true
		errorTask := task.New("errorer", errorMe)
		serialisedError, err := taskSerialiser.Serialise(errorTask)
		require.NoError(t, err)

		// Act
		_, err = client.LPush(ctx, "tasks", serialisedError).Result()
		require.NoError(t, err)

		redisResult, err := client.BRPop(ctx, 5*time.Second, "results").Result()
		require.NoError(t, err)

		errorResult, err := resultSerialiser.Deserialise([]byte(redisResult[1]))
		require.NoError(t, err)

		// Assert
		expectedError := fmt.Errorf("error for payload: %v", errorMe)
		assert.Equal(t, expectedError.Error(), errorResult.ErrMsg)
		assert.Equal(t, errorTask.ID, errorResult.TaskID)
	})
}
