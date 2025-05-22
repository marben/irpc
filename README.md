# irpc - interface based rpc

Irpc is a library and a code generator that generates network code based on your go interface definition.

Irpc takes your interface definition and turns it into network client/server code, that implements/uses your interface. Allowing for seamless dependency injection into your code.

### example network api definition:
```go
// code inside math.go
type Math interface {
    Add(a, b int) (int, error)
}
```

running `$ irpc math.go` generates `math_irpc.go` with client/service definition 

### generated service
returns service, that takes `Math` interface implementation and forward incoming network requests to it, returning the results back over network
```go
func NewMathIRpcService(impl Math) *MathIRpcService
```

### generated client
returns a working client, that implements Math interface, meaning we can directly call Add(1,2) on it
```go
func NewMathIRpcClient(endpoint *irpc.Endpoint) (*MathIRpcClient, error) {...}


client, err := NewMathIRpcClient(ep)
result, err := client.Add(1,2) // a network call
fmt.Println(result)    // "3"  (presuming our implementation did a simple addition)
```

and turns it into client and server code, that can be used just as the interface, but over a network

# installation
- import into your project module: `go get github.com/marben/irpc`
- install binary to generate code: `go install github.com/marben/cmd/irpc@latest`