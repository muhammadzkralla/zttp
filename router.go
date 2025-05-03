package zttp

import "strings"

type Handler func(req Req, res Res)

type Route struct {
	path    string
	handler Handler
}

type Router struct {
	getRoutes    []Route
	postRoutes   []Route
	deleteRoutes []Route
	putRoutes    []Route
	patchRoutes  []Route
	middlewares  []MiddlewareWrapper
}

// Register the passed handler and path with the app's get routes
func (app *App) Get(path string, handler Handler) {
	app.getRoutes = append(app.getRoutes, Route{path, applyMiddleware(handler, app)})
}

// Register the passed handler and path with the app's delete routes
func (app *App) Delete(path string, handler Handler) {
	app.deleteRoutes = append(app.deleteRoutes, Route{path, applyMiddleware(handler, app)})
}

// Register the passed handler and path with the app's post routes
func (app *App) Post(path string, handler Handler) {
	app.postRoutes = append(app.postRoutes, Route{path, applyMiddleware(handler, app)})
}

// Register the passed handler and path with the app's put routes
func (app *App) Put(path string, handler Handler) {
	app.putRoutes = append(app.putRoutes, Route{path, applyMiddleware(handler, app)})
}

// Register the passed handler and path with the app's patch routes
func (app *App) Patch(path string, handler Handler) {
	app.patchRoutes = append(app.patchRoutes, Route{path, applyMiddleware(handler, app)})
}

// This function searches for the matching handler for the passed request path
// As well as extracting the params, if exist
func matchRoute(requestPath string, routes []Route) (Handler, map[string]string) {
	for _, route := range routes {
		params := make(map[string]string)

		// Split both route and request paths with `/` delimiter to compare each part respectively
		routeParts := strings.Split(route.path, "/")
		requestParts := strings.Split(requestPath, "/")

		// if the lengths don't match, we don't have to compare to know they don't match
		if len(routeParts) != len(requestParts) {
			continue
		}

		match := true
		for i := range routeParts {
			// If it's a param, parse it
			// Otherwise, check for equality
			if strings.HasPrefix(routeParts[i], ":") {
				paramName := routeParts[i][1:]
				params[paramName] = requestParts[i]
			} else if routeParts[i] != requestParts[i] {
				match = false
				break
			}
		}

		// If matched, return it
		if match {
			return route.handler, params
		}
	}

	// Default: no match
	return nil, nil
}
