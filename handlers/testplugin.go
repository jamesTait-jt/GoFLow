package main

import (
	"fmt"
	"time"

	"github.com/jamesTait-jt/goflow/pkg/task"
	"golang.org/x/exp/rand"
)

func handle(payload any) task.Result {
	fmt.Println("Handling task with payload: ", payload)

	rand.Seed(uint64(time.Now().UnixNano()))
	n := rand.Intn(1000)
	fmt.Printf("Sleeping %d milliseconds...\n", n)
	time.Sleep(time.Millisecond * time.Duration(n))

	fmt.Println("Done")

	return task.Result{Payload: "Success!!"}
}

func NewHandler() task.Handler {
	return handle
}
