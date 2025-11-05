package autorpc

// Both Server and Group implement this interface.
type Registerer interface {
	// register is called by RegisterMethod to register a method.
	register(name string, fn interface{}, allMiddlewares *MiddlewareChain)
}
