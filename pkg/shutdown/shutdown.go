package shutdown

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/jamesTait-jt/goflow/pkg/log"
)

func AddShutdownHook(ctx context.Context, logger log.Logger, closers ...io.Closer) {
	c := make(chan os.Signal, 1)
	signal.Notify(
		c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM,
	)

	select {
	case <-c:
		logger.Info("received termination signal, initiating graceful shutdown...")
	case <-ctx.Done():
		logger.Info("context canceled, initiating graceful shutdown...")
	}

	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			logger.Error(fmt.Sprintf("failed to stop closer: %v", err))
		}
	}

	logger.Info("completed graceful shutdown")
}
