package autorpc

type Group struct {
	server      *Server
	prefix      string
	middlewares *MiddlewareChain
}

// Group creates a new group with the given prefix.
// The prefix will be prepended to all method names registered in this group.
// An empty prefix is allowed.
// Middleware can be provided as variadic parameters and will apply to all methods in the group.
//
// Example:
//
//	mathGroup := server.Group("math.", TimingMiddleware, AuthMiddleware())
func (s *Server) Group(prefix string, middlewares ...Middleware) *Group {
	group := &Group{
		server:      s,
		prefix:      prefix,
		middlewares: NewMiddlewareChain(middlewares...),
	}
	return group
}

// Use adds middleware to the group.
// Middleware added to a group will apply to all methods registered in that group.
func (g *Group) Use(middlewares ...Middleware) {
	for _, mw := range middlewares {
		g.middlewares.Add(mw)
	}
}

// register implements the Registerer interface for Group.
// It combines the group prefix with the method name and combines
// group middlewares with method-specific middlewares before delegating to the server.
func (g *Group) register(name string, fn interface{}, methodMiddlewares *MiddlewareChain) {
	fullName := g.prefix + name
	allMiddlewares := NewMiddlewareChain()

	if g.middlewares != nil && g.middlewares.Len() > 0 {
		for i := 0; i < g.middlewares.Len(); i++ {
			allMiddlewares.Add(g.middlewares.middlewares[i])
		}
	}

	if methodMiddlewares != nil && methodMiddlewares.Len() > 0 {
		for i := 0; i < methodMiddlewares.Len(); i++ {
			allMiddlewares.Add(methodMiddlewares.middlewares[i])
		}
	}

	g.server.register(fullName, fn, allMiddlewares)
}
