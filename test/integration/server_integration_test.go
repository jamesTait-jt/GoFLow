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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func Test_Server_Integration(t *testing.T) {
	// Arrange
	ctx := context.Background()

	redisContainer, err := startRedisContainer(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		testcontainers.CleanupContainer(t, redisContainer)
	})

	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	redisClient, err := connectToRedisContainer(ctx, endpoint)
	require.NoError(t, err)
	t.Cleanup(func() {
		redisClient.Close()
	})

	serverConfig := &config.Config{
		BrokerType: "redis",
		BrokerAddr: endpoint,
	}

	serverRuntime := runtime.Runtime{
		Conf: serverConfig,
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		runtimeErr := serverRuntime.Run(ctx)
		require.NoError(t, runtimeErr)
	}()

	logger := log.NewConsoleLogger()
	serverAddr := fmt.Sprintf(":%d", 50051)

	goFlowService, err := client.NewGoFlowService(serverAddr, time.Minute, logger)
	require.NoError(t, err)

	t.Run("", func(t *testing.T) {
		// Arrange
		// Act
		_, err := goFlowService.Push("test-type", "payload")

		// Assert
		assert.NoError(t, err)
	})
}
