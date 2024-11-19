package middleware

import (
	"context"
	"time"

	"github.com/jamesTait-jt/goflow/task"
	"github.com/sirupsen/logrus"
)

func Retry(next task.Handler, tries int) task.Handler {
	return func(payload any) task.Result {
		var result task.Result
		for i := 0; i < tries; i++ {
			result = next(payload)
			if result.ErrMsg == "" {
				return result
			}

			logrus.WithField("attempts", i+1).Warn("Retrying task...")
		}

		return result
	}
}

func Timeout(next task.Handler, timeout time.Duration) task.Handler {
	return func(payload any) task.Result {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		resultChan := make(chan task.Result, 1)
		go func() {
			result := next(payload)
			resultChan <- result
		}()

		select {
		case result := <-resultChan:
			return result
		case <-ctx.Done():
			return task.Result{ErrMsg: "Task timed out"}
		}
	}
}

func ReportTime(next task.Handler) task.Handler {
	return func(payload any) task.Result {
		start := time.Now()
		result := next(payload)
		duration := time.Since(start)

		logrus.WithFields(logrus.Fields{
			"duration": duration,
			"success":  result.ErrMsg == "",
		}).Info("Task processed")

		return result
	}
}
