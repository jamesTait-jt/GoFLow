//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jamesTait-jt/goflow/cmd/server/config"
	"github.com/jamesTait-jt/goflow/cmd/server/runtime"
	"github.com/jamesTait-jt/goflow/grpc/client"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/pkg/serialise"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func Test_Server_Integration(t *testing.T) {
	// Arrange
	ctx := context.Background()

	redisContainer, err := startRedisContainer(ctx)
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, redisContainer)

	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	redisClient, err := connectToRedisContainer(ctx, endpoint)
	require.NoError(t, err)
	defer redisClient.Close()

	taskSerialiser := serialise.NewGobSerialiser[task.Task]()
	resultSerialiser := serialise.NewGobSerialiser[task.Result]()

	serverConfig := &config.Config{
		BrokerType: "redis",
		BrokerAddr: endpoint,
	}

	serverRuntime := runtime.Runtime{
		Conf: serverConfig,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		runtimeErr := serverRuntime.Run(ctx)
		require.NoError(t, runtimeErr)
	}()

	logger := log.NewConsoleLogger()
	serverAddr := fmt.Sprintf(":%d", 50051)

	goFlowService, err := client.NewGoFlowService(serverAddr, time.Minute, logger)
	require.NoError(t, err)

	t.Run("Handles Push requests and forwards them to tasks queue", func(t *testing.T) {
		// Arrange
		taskType := "test-type"
		payload := `{"value": "20"}`

		// Act
		id, err := goFlowService.Push(taskType, payload)
		require.NoError(t, err)

		redisResult, err := redisClient.BRPop(ctx, 5*time.Second, "tasks").Result()
		require.NoError(t, err)

		tsk, err := taskSerialiser.Deserialise([]byte(redisResult[1]))
		require.NoError(t, err)

		// Assert
		expectedTask := task.Task{
			ID:      id,
			Type:    taskType,
			Payload: payload,
		}

		assert.Equal(t, expectedTask, tsk)
	})

	t.Run("Reads results from the results queue and returns them via Get requests", func(t *testing.T) {
		// Arrange
		taskID := "12345"
		resultPayload := "result payload"
		result := task.Result{
			TaskID:  taskID,
			Payload: resultPayload,
		}
		serialisedResult, err := resultSerialiser.Serialise(result)
		require.NoError(t, err)

		// Act
		_, err = redisClient.LPush(ctx, "results", serialisedResult).Result()
		require.NoError(t, err)

		returnedResult, err := goFlowService.Get(taskID)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, "result payload", returnedResult)
	})

	t.Run("Reads error results from the results queue and returns them via Get requests", func(t *testing.T) {
		// Arrange
		taskID := "12345"
		resultError := "ERROR: result error"
		result := task.Result{
			TaskID: taskID,
			ErrMsg: resultError,
		}
		serialisedResult, err := resultSerialiser.Serialise(result)
		require.NoError(t, err)

		// Act
		_, err = redisClient.LPush(ctx, "results", serialisedResult).Result()
		require.NoError(t, err)

		returnedResult, err := goFlowService.Get(taskID)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, "ERROR: result error", returnedResult)
	})
}
