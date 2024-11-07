package main

import "github.com/jamesTait-jt/goflow/cmd/workerpool/runtime"

func main() {
	r := runtime.New()
	r.Run()
}
