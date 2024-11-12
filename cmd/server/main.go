package main

import (
	"log"

	"github.com/jamesTait-jt/goflow/cmd/server/runtime"
)

func main() {
	r := runtime.New()

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
