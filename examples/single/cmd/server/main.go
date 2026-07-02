package main

import (
	"fmt"
	"os"

	"github.com/marben/irpc"

	"github.com/marben/irpc/examples/single"
)

type Sample struct{}

func (s *Sample) Greeting(name string) string {
	return fmt.Sprintf("Hello, %s", name)
}

func main() {
	service := single.NewSampleIrpcService(&Sample{})
	handler := irpc.NewSingleHandler(os.Stdin, os.Stdout, service)
	for {
		if err := handler.HandleOnce(); err != nil {
			print(err.Error())
			os.Exit(1)
		}
	}
}
