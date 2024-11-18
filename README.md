# goflow - A Distributed Task Queue for Go

[![Build Status](https://github.com/jamesTait-jt/goflow/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/jamesTait-jt/goflow/actions/workflows/main.yml)
[![Go Report Card](https://goreportcard.com/badge/jamesTait-jt/goflow)](https://goreportcard.com/report/jamesTait-jt/goflow)
[![codecov](https://codecov.io/github/jamesTait-jt/goflow/branch/main/graph/badge.svg?token=JW9HOXRPJ1)](https://codecov.io/github/jamesTait-jt/goflow)

GoFlow is a scalable and flexible framework for task orchestration, supporting both local and distributed execution modes. It enables efficient task processing through worker pools, customizable task handlers, and pluggable brokers for seamless task and result communication.

There are three ways to utilize GoFlow, catering to different levels of complexity:

1. **Local Mode (as a library)**
    The simplest method, integrating GoFlow directly into your Go application. This sets up an in-process worker pool, which you interact with via the GoFlow object to send tasks and retrieve results.

2. **Distributed Mode (as a library)**
    Deploy GoFlow in a distributed setup by using a message broker to mediate communication between your application and worker pools. Your application sends tasks and receives results via the GoFlow object, while the worker pools are configured to interact with your chosen message broker.

3. **Distributed Mode (via gRPC client)**
    Deploy GoFlow on a Kubernetes cluster using the CLI. In this mode, you interact with GoFlow using a lightweight gRPC client, eliminating the need to include the full GoFlow library in your application.

## User Guide

### Prerequisites

Go (>=1.21)

### Local Mode

To use GoFlow as an embedded process in your application, install it via go get:

```bash
go get github.com/jamesTait-jt/goflow
```

This will add GoFlow as a dependency to your project. You can then create a GoFlow object in your code:

```go
// Create an in memory store to keep track of custom handlers
taskHandlerStore := store.NewInMemoryKVStore[string, task.Handler]()

// Inject the store into the GoFlow object
gf := goflow.NewLocalMode(taskHandlerStore)
```

By default, this initializes the GoFlow object with standard configuration (see options.go). You can customize these settings using functional options, such as `WithResultsStore` to define a custom results store or `WithNumWorkers` to set the number of worker goroutines. Once configured, the GoFlow object is ready to start:

```go
if err := gf.Start(); err != nil {
    // Handle the error
}
```

This starts the worker pool to process tasks. When you're finished, gracefully shut down GoFlow using the Close method:

```go
if err := gf.Close(); err != nil {
    // Handle the error
}
```

This stops the worker pool and closes any open resources.

#### Task handlers

In GoFlow, task handlers are functions that process tasks submitted to the framework. A task handler takes a payload (of type any) and returns a task.Result, which contains the result of processing the task. Task handlers are registered to specific task types, allowing GoFlow to route tasks to the appropriate handler when processed.

Below is an example that demonstrates how to define and register a task handler in GoFlow. The task handler will copy the payload sent on the task into the result payload:

```go
repeater := func(payload any) task.Result {
    return task.Result{Payload: payload)}
}
```

To ensure GoFlow uses the correct handler for a given task type, we register the handler with a specific task type. In this case, we register the handler for the `repeater` task type:

```go
taskType := "repeater"
gf.RegisterHandler(taskType, repeater)
```

Handlers can be registered either before starting GoFlow or dynamically while it is running.

Once the handler is registered, we can push tasks to GoFlow for processing. Each task will be picked up by a worker, which will retrieve the appropriate handler from the registry to process the task.

```go
taskID, err := gf.Push("repeater", "Hello, GoFlow!")
if err != nil {
    // Handle error
}
```

GoFlow assigns a unique task ID and returns it so you can later retrieve the result.

```go
result, ok, err := gf.GetResult(taskIDs[i])
if err != nil {
    // Handle error
}

if !ok {
    // Task did not have a corresponding result (it may not be finished)
}
```

#### Task handler store & results store

You can define custom task handler registries and result stores by implementing the KVStore interface. For instance, you might want to persist task results in a database or a cloud service.

Note that the task handler registry must reside in memory, as functions are not serialisable.

### Configuration

copy your handlers into minikube

### Usage
