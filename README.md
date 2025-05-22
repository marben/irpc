# IRPC - interface based rpc generator for go programming language

IRPC is a library and code generator that generates network code based on your go interface definition.

It aims to make the network code as invisible as possible, allowing you to treat the generated client code as any other implementation of your interface. No wrappers needed.

IRPC is very efficient as it doesn't transfer metadata. Payload layout is defined in the generated code, which is used by both client and server. Layout definition/api version is verified upon network handshake.

IRPC allows for bidirectional rpc calls over any io.ReadWriteCloser such as Tcp socket or Websocket etc

### Example network api definition:
```go
// api definition inside math.go
type Math interface {
    Add(a, b int) (int, error)
}
```

Running IRPC command tool`irpc math.go` generates file `math_irpc.go` with client and service implementation.

### Generated client
A working client, that implements Math interface, meaning we can directly call Add(1,2) on it
```go
// inside math_irpc.go
func NewMathIRpcClient(endpoint *irpc.Endpoint) (*MathIRpcClient, error) {...}
```
```go
// example use:
client, err := NewMathIRpcClient(ep)
result, err := client.Add(1,2) // a network call
fmt.Println(result)    // "3"  (presuming our implementation did a simple addition)
```

### Generated service
Service that takes `Math` interface implementation and forward incoming network requests to it, returning the results back over network
```go
// inside math_irpc.go
func NewMathIRpcService(impl Math) *MathIRpcService
```


# Installation
IRPC code generator tool is installed by:  
`go install github.com/marben/irpc/cmd/irpc@latest`  

You may need to add go bin directory (typically `~/go/bin`) to your `PATH` env variable.  
IRPC command tool is called `irpc`

IRPC library needs to be added as dependency to your go module:  
`go get github.com/marben/irpc`
```go
// import
import "github.com/marben/irpc"

func main() {
	// irpc package name is 'irpc'
	srvr, err := irpc.NewServer(...)
}
```

# Working example
Have a look to see, how easy it is:  
[github.com/marben/irpc_tcp_example](https://github.com/marben/irpc_tcp_example) - easy to read example code