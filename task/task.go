package task

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Handler processes task payloads
type Handler func(payload any) Result

// Type represents a generic task structure
type Task struct {
	ID      string
	Type    string
	Payload any
	Opts    taskOption
}

type taskOption struct {
	Retries     int
	LogDuration bool
	Timeout     time.Duration
}

type Result struct {
	TaskID  string
	Payload any
	ErrMsg  string
}

type TaskOption interface {
	apply(*taskOption)
}

func New(taskType string, payload any, opts ...TaskOption) Task {
	id := uuid.New()
	t := Task{
		ID:      id.String(),
		Type:    taskType,
		Payload: payload,
	}

	return t
}

// nolint:revive // stuttering here is acceptable
type TaskOrResult interface {
	Task | Result
}

type Submitter[T TaskOrResult] interface {
	Submit(ctx context.Context, t T) error
}

type Dequeuer[T TaskOrResult] interface {
	Dequeue(ctx context.Context) <-chan T
}
