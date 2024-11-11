package main

import (
	"fmt"
	"time"

	"github.com/jamesTait-jt/goflow"
	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/pkg/serialise"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("starting")

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	fmt.Println("conntected to redis")

	resultsStore := store.NewInMemoryKVStore[string, task.Result]()

	taskSerialiser := serialise.NewGobSerialiser[task.Task]()
	resultSerialiser := serialise.NewGobSerialiser[task.Result]()
	logger := log.NewConsoleLogger()

	gf := goflow.New(
		broker.NewRedisBroker[task.Task](redisClient, "tasks", taskSerialiser, nil, logger),
		broker.NewRedisBroker[task.Result](redisClient, "results", nil, resultSerialiser, logger),
		goflow.WithResultsStore(resultsStore),
	)

	gf.Start()

	maxItrs := 100

	results := make(chan task.Result)

	for i := 0; i < maxItrs; i++ {
		id, _ := gf.Push("testplugin", "Im a random sleeper")

		go func() {
			for {
				result, ok := gf.GetResult(id)
				if !ok {
					time.Sleep(time.Second)
					continue
				}

				results <- result
				if i == maxItrs-1 {
					close(results)
				}

				return
			}
		}()
	}

	i := 0
	for r := range results {
		i++
		fmt.Println(i, r)
	}
}
