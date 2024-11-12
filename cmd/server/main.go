package main

import (
	"context"
	"log"

	"github.com/jamesTait-jt/goflow/cmd/server/runtime"
)

func main() {
	r := runtime.New()

	ctx := context.Background()
	if err := r.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
