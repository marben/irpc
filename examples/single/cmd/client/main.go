package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/marben/irpc"

	"github.com/marben/irpc/examples/single"
)

func main() {
	proc, err := Proc("go", "run", "./examples/single/cmd/server/main.go")
	if err != nil {
		log.Fatal(err)
	}
	defer proc.Close()
	go io.Copy(os.Stderr, proc.Stderr())
	ep := irpc.NewEndpoint(proc)
	defer ep.Close()
	client, err := single.NewSampleIrpcClient(ep)
	if err != nil {
		log.Fatal(err)
	}
	for i := range 100000 {
		fmt.Println(client.Greeting(fmt.Sprintf("Hello%d", i)))
	}
}
