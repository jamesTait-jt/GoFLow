package main

import (
	"log"

	"github.com/jamesTait-jt/goflow/cmd/workerpool/runtime"
)

func main() {
	r := runtime.New()

	if err := r.Run(); err != nil {
		log.Fatal(err.Error())
	}
}
