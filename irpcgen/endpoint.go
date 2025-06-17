package irpcgen

import "context"

type Endpoint interface {
	// RegisterClient registers the client with the server.
	RegisterClient(serviceId []byte) error
	// CallRemoteFunc calls a remote function on the server.
	CallRemoteFunc(ctx context.Context, serviceId []byte, funcId FuncId, params Serializable, resp Deserializable) error
}
