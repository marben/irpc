package irpcgen

import "context"

// Endpoint represents one side of an active RPC connection.
// Each Endpoint communicates with one peer Endpoint on the other side.
type Endpoint interface {
	// RegisterClient registers a client with the peer Endpoint.
	RegisterClient(serviceId []byte) error
	// CallRemoteFunc invokes a function on the peer Endpoint.
	CallRemoteFunc(ctx context.Context, serviceId []byte, funcId FuncId, params Serializable, resp Deserializable) error
}
