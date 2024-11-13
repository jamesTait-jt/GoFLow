//go:build integration

package integration

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

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
