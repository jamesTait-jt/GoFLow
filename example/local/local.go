package main

import (
	"fmt"
	"time"

	"github.com/jamesTait-jt/goflow"
	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/jamesTait-jt/goflow/workerpool"
	"golang.org/x/exp/rand"
)

var (
	numWorkers = 10
	queueSize  = 5
)

func main() {
	resultsStore := store.NewInMemoryKVStore[string, task.Result]()
	channelBroker := broker.NewChannelBroker[task.Task](queueSize)
	taskHandlerStore := store.NewInMemoryKVStore[string, task.Handler]()
	workerPool := workerpool.New(numWorkers)

	gf := goflow.New(
		goflow.WithLocalMode(workerPool, taskHandlerStore),
		goflow.WithResultsStore(resultsStore),
		goflow.WithTaskBroker(channelBroker),
	)

	taskHandler := func(payload any) task.Result {
		rand.Seed(uint64(time.Now().UnixNano()))
		n := rand.Intn(1000)
		fmt.Printf("Sleeping %d milliseconds...\n", n)
		time.Sleep(time.Millisecond * time.Duration(n))

		return task.Result{Payload: fmt.Sprintf("Processed: %v", payload)}
	}

	taskType := "exampleTask"
	gf.RegisterHandler(taskType, taskHandler)

	gf.Start()

	taskIDs := []string{}

	for i := 0; i < 100; i++ {
		taskID, err := gf.Push(taskType, "My example payload")
		if err != nil {
			fmt.Printf("Error pushing task: %v\n", err)
			return
		}

		fmt.Printf("Task submitted with ID: %s\n", taskID)
		taskIDs = append(taskIDs, taskID)
	}

	time.Sleep(time.Second * 1)

	for i := 0; i < len(taskIDs); i++ {
		result, _ := gf.GetResult(taskIDs[i])
		fmt.Println(result)
	}

	gf.Stop()
}
