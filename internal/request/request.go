package request

// Request defines a single API request. The engine layer (transport, protocol
// dispatch, authentication, variable interpolation, scripting) builds on this
// definition.
//
// The transport remains abstract so additional protocols (WebSocket, gRPC,
// SSE, MQTT, ...) can be added without changing the application architecture.
type Request struct {
	ID     string
	Name   string
	Method Method
	URL    string

	Headers []Header
	Query   []Parameter
	Body    string

	Auth    Auth
	Timeout int64 // milliseconds; 0 means "no explicit timeout"
}

// Method is an HTTP method.
type Method string

const (
	MethodGet     Method = "GET"
	MethodPost    Method = "POST"
	MethodPut     Method = "PUT"
	MethodPatch   Method = "PATCH"
	MethodDelete  Method = "DELETE"
	MethodHead    Method = "HEAD"
	MethodOptions Method = "OPTIONS"
)

// Header is a single request header.
type Header struct {
	Key   string
	Value string
}

// Parameter is a query or path parameter.
type Parameter struct {
	Key   string
	Value string
}

// Auth describes the authentication configuration attached to a request.
// Implementations live in the auth package and are dispatched by type.
type Auth struct {
	Type   string
	Config map[string]string
}
