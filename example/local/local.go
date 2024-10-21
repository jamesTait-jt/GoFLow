package main

import (
	"fmt"
	"math/big"
	"time"

	"crypto/rand"

	"github.com/jamesTait-jt/goflow"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
)

var (
	numWorkers                  = 10
	queueSize                   = 5
	maxRandomSleepInMlliseconds = 1000
)

func main() {
	resultsStore := store.NewInMemoryKVStore[string, task.Result]()
	taskHandlerStore := store.NewInMemoryKVStore[string, task.Handler]()

	gf := goflow.NewLocalMode(
		numWorkers, queueSize, queueSize,
		taskHandlerStore,
		goflow.WithResultsStore(resultsStore),
	)

	taskHandler := func(payload any) task.Result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(maxRandomSleepInMlliseconds)))
		fmt.Printf("Sleeping %d milliseconds...\n", n.Int64())
		time.Sleep(time.Millisecond * time.Duration(n.Int64()))

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
