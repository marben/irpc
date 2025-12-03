# irpc — Interface-based RPC generator for Go

irpc is a lightweight RPC framework for Go that generates client and server implementations directly from Go interfaces.  
You implement your interface once, and irpc generates the network glue: encoding, decoding, dispatch, and a type-safe client.

**No wrappers. No boilerplate. Just interfaces.**

---

## ✨ Features

- **Interface-first design** — define a Go interface, and irpc generates:
  - A server implementation that exposes it over a connection
  - A client that implements the same interface and calls the server remotely
- **Zero JSON / reflection overhead** — compact binary encoding with versioned payload metadata
- **Fast startup** — minimal handshake, no schema registry, no IDL
- **Type-safe:** client has the exact same method signatures as your interface
- **Transport-agnostic:** works with any `io.ReadWriteCloser` (TCP, WebSocket, pipes, process stdio…)
- **Small and dependency-free** — no heavy runtime

---

## ⚠️ Status

**Experimental (v0.x).**  
APIs may change until v1. Feedback and contributions are welcome.

---

## 🚀 Quickstart

### 1. Install the generator

```bash
go install github.com/marben/irpc/cmd/irpc@latest
```
### 2. Define a Go interface
``` go
// math.go
package mathsvc

type Math interface {
    Add(a, b int) (int, error)
    Mul(a, b int) (int, error)
}
```

### 3. Generate RPC code
```bash
irpc math.go
```
This produces math_irpc.go with:

```go
NewMathIrpcClient(endpoint irpcgen.Endpoint) (*MathIrpcClient, error)
```
where MathIrpcClient implements Math interface

```go
NewMathIrpcService(impl Math) *MathIrpcService
```

### 4. Use it — Server
```go
ln, _ := net.Listen("tcp", ":9000")
```

### 5. Use it — Client
```go
res, err := cli.Add(2, 3)
fmt.Println(res) // => 5
```

📦 Example Project

A working example with client/server code is here:

👉 https://github.com/marben/irpc-examples


(including build scripts and generated code)

📡 Protocol


📄 License

MIT — see LICENSE

🤝 Contributing

PRs, issues, and feedback are welcome.

**old readme**
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
func NewMathIrpcClient(endpoint *irpc.Endpoint) (*MathIrpcClient, error) {...}
```
```go
// example use:
client, err := NewMathIrpcClient(ep)
result, err := client.Add(1,2) // a network call
fmt.Println(result)    // "3"  (presuming our implementation did a simple addition)
```

### Generated service
Service that takes `Math` interface implementation and forward incoming network requests to it, returning the results back over network
```go
// inside math_irpc.go
func NewMathIrpcService(impl Math) *MathIrpcService
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
