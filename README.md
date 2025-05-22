# IRPC - interface based rpc generator for go programming language

IRPC is a library and a code generator that generates network code based on your go interface definition.

IRPC takes your interface definition and turns it into network client/server code, that implements/uses your interface. Allowing for seamless use of generated network client code in place of the defining interface.

### example network api definition:
```go
// definition inside math.go
type Math interface {
    Add(a, b int) (int, error)
}
```

Running IRPC command tool`$ irpc math.go` generates `math_irpc.go` with client/service definition.

### generated client
returns a working client, that implements Math interface, meaning we can directly call Add(1,2) on it
```go
func NewMathIRpcClient(endpoint *irpc.Endpoint) (*MathIRpcClient, error) {...}


client, err := NewMathIRpcClient(ep)
result, err := client.Add(1,2) // a network call
fmt.Println(result)    // "3"  (presuming our implementation did a simple addition)
```

### generated service
returns service, that takes `Math` interface implementation and forward incoming network requests to it, returning the results back over network
```go
// generated service inside math_irpc.go
func NewMathIRpcService(impl Math) *MathIRpcService
```


# Installation
IRPC code generator tool is installed by:`go install github.com/marben/irpc/cmd/irpc@latest`  
You may need to add go bin directory (typically `~/go/bin`) to your `PATH` env variable.  
Command tool is called `irpc`

IRPC library needs to be added as dependency to your go module:`go get github.com/marben/irpc`
```go
// import
import "github.com/marben/irpc"

func main() {
	// irpc package name is 'irpc'
	srvr, err := irpc.NewServer(...)
}
```