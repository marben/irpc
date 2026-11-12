package irpcgen

import "context"

// Endpoint represents one side of an active RPC connection.
// Each Endpoint communicates with one peer Endpoint on the other side.
type Endpoint interface {
	// RegisterClient registers a client with the peer Endpoint.
	RegisterClient(serviceId ServiceId) error
	// CallRemoteFunc invokes a function on the peer Endpoint.
	CallRemoteFunc(ctx context.Context, serviceId ServiceId, funcId FuncId, params Serializable, resp Deserializable) error
}

// UnlimitedFuncCaller is optionally implemented by an [Endpoint].
//
// CallRemoteFuncUnlimited works like CallRemoteFunc, but the call is not counted against
// the endpoint's limit of parallel calls. Generated code uses it for calls that can last for
// a long time and that other calls depend on (streams), which would otherwise deadlock.
type UnlimitedFuncCaller interface {
	CallRemoteFuncUnlimited(ctx context.Context, serviceId ServiceId, funcId FuncId, params Serializable, resp Deserializable) error
}
