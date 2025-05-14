# irpc - interface based rpc

a tool that generates network bridge code for you interface

a simple command line tool and library that takes your interface definition
such as:
```go
type Greeter interface {
	Greet(msg string) (resp string, err error)
}
```

and turns it into client and server code, that can be used just as the interface, but over a network

# installation
- binary: `go install github.com/marben/cmd/irpc@latest`
- import into project: `go get github.com/marben/irpc/pkg/irpc`