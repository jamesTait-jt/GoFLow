package main

import (
	"log"

	"github.com/jamesTait-jt/goflow/cmd/goflow/runtime"
)

func main() {
	r := runtime.New()

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
