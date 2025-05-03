package zttp

type Middleware func(req Req, res Res, next func())

// MiddlewareWrapper wraps a Middleware with a certain path
// If the path is empty, the middleware will be applied globally to all requests
// If the path is set, the middleware will only be applied to matched requests
type MiddlewareWrapper struct {
	Path       string
	Middleware Middleware
}

// Use registers a middleware function with the app
// Example1: app.Use(func(req, res, next) { ... })
// Example2: app.Use("/admin", func(req, res, next) { ... })
// Otherwise it will panic, don't panic her, please :)
func (app *App) Use(args ...any) {
	path := ""
	var middleware Middleware

	// If only one arg is passed, it's expected to be the middleware itself
	// Otherwise, the first arg is expected to be the path and the other is the middleware
	if len(args) == 1 {
		m, ok := args[0].(func(Req, Res, func()))
		if !ok {
			panic("Invalid argument: expected handler function")
		}

		middleware = m
	} else if len(args) == 2 {
		p, ok1 := args[0].(string)
		m, ok2 := args[1].(func(Req, Res, func()))
		if !ok1 || !ok2 {
			panic("Invalid arguments: expected string path and handler function")
		}

		path = p
		middleware = m
	} else {
		panic("Invalid argument: expected handler function")
	}

	// Register the middleware with the app middlewares
	mw := MiddlewareWrapper{
		Path:       path,
		Middleware: middleware,
	}
	app.middlewares = append(app.middlewares, mw)
}

// This function constructs a chain of functions to be called one after the other
func applyMiddleware(finalHandler Handler, router *Router) Handler {

	// The req and res arguments are the ones created in the server file
	return func(req Req, res Res) {

		// Store the index of the current middleware globally before incrementing it recursively
		currentMiddlewareIdx := 0

		var next func()
		next = func() {

			// If there are still middlewares to check in the app's middlewares
			// Combine global middlewares and router-specific middlewares
			allMiddlewares := append(app.middlewares, router.middlewares...)
			if currentMiddlewareIdx < len(allMiddlewares) {
				middlewareWrapper := allMiddlewares[currentMiddlewareIdx]

				// Increment the middleware index to check the next middleware in the next
				// recursive call
				currentMiddlewareIdx++

				// If the middleware path matched or it's a global middleware, execute it
				// Otherwise, call the next middleware
				if middlewareWrapper.Path == "" || middlewareWrapper.Path == req.Path {
					middlewareWrapper.Middleware(req, res, next)
				} else {
					next()
				}
			} else {
				// After all the middlewares have been processed, call the actual request handler
				finalHandler(req, res)
			}
		}

		// Start the chain reaction or smth
		// Note that the next() function now holds the calls to the matched middlewares first
		// and finally the call to the request handler itself
		// So, when we call next() here, we literally start the chained function calls
		next()
	}
}
