package zttp

import (
	"log"
	"net"
	"path"
	"strings"
)

type Handler func(req *Req, res *Res)

type Route struct {
	path    string
	handler Handler
}

type Router struct {
	*App
	prefix       string
	getRoutes    []Route
	postRoutes   []Route
	deleteRoutes []Route
	putRoutes    []Route
	patchRoutes  []Route
	middlewares  []MiddlewareWrapper
}

// Register the passed handler and path with the app's get routes
func (app *App) Get(path string, handler Handler) {
	app.getRoutes = append(app.getRoutes, Route{path, applyMiddleware(handler, app.Router)})
}

// Register the passed handler and path with the app's delete routes
func (app *App) Delete(path string, handler Handler) {
	app.deleteRoutes = append(app.deleteRoutes, Route{path, applyMiddleware(handler, app.Router)})
}

// Register the passed handler and path with the app's post routes
func (app *App) Post(path string, handler Handler) {
	app.postRoutes = append(app.postRoutes, Route{path, applyMiddleware(handler, app.Router)})
}

// Register the passed handler and path with the app's put routes
func (app *App) Put(path string, handler Handler) {
	app.putRoutes = append(app.putRoutes, Route{path, applyMiddleware(handler, app.Router)})
}

// Register the passed handler and path with the app's patch routes
func (app *App) Patch(path string, handler Handler) {
	app.patchRoutes = append(app.patchRoutes, Route{path, applyMiddleware(handler, app.Router)})
}

// Register the passed handler and path with the router's get routes
func (router *Router) Get(path string, handler Handler) {
	router.getRoutes = append(router.getRoutes, Route{cleanPath(router.prefix, path), applyMiddleware(handler, router)})
}

// Register the passed handler and path with the router's delete routes
func (router *Router) Delete(path string, handler Handler) {
	router.deleteRoutes = append(router.deleteRoutes, Route{cleanPath(router.prefix, path), applyMiddleware(handler, router)})
}

// Register the passed handler and path with the router's post routes
func (router *Router) Post(path string, handler Handler) {
	router.postRoutes = append(router.postRoutes, Route{cleanPath(router.prefix, path), applyMiddleware(handler, router)})
}

// Register the passed handler and path with the router's put routes
func (router *Router) Put(path string, handler Handler) {
	router.putRoutes = append(router.putRoutes, Route{cleanPath(router.prefix, path), applyMiddleware(handler, router)})
}

// Register the passed handler and path with the router's patch routes
func (router *Router) Patch(path string, handler Handler) {
	router.patchRoutes = append(router.patchRoutes, Route{cleanPath(router.prefix, path), applyMiddleware(handler, router)})
}

func cleanPath(prefix, p string) string {
	// Ensure prefix starts with "/" and does not end with "/"
	if prefix == "" {
		prefix = "/"
	}
	if prefix != "/" {
		prefix = path.Clean("/" + prefix)
	}

	// Ensure path starts with "/"
	if !path.IsAbs(p) {
		p = "/" + p
	}
	p = path.Clean(p)

	// Concatenate and clean the final path
	full := path.Join(prefix, p)
	if full != "/" && full[len(full)-1] == '/' {
		full = full[:len(full)-1]
	}
	return full
}

// Find the matched handler with the passed path from the router and parse params, if exist
func findHandler(method, path string, socket net.Conn, app *App) (Handler, map[string]string) {
	for _, router := range app.Routers {
		var routes []Route
		switch method {
		case "GET":
			routes = router.getRoutes
		case "DELETE":
			routes = router.deleteRoutes
		case "POST":
			routes = router.postRoutes
		case "PUT":
			routes = router.putRoutes
		case "PATCH":
			routes = router.patchRoutes
		default:
			log.Println("unsupported method:", method)
			sendResponse(socket, []byte("Method Not Allowed"), 405, "text/plain", nil)
			return nil, nil
		}

		if handler, params := matchRoute(path, routes); handler != nil {
			return handler, params
		}
	}

	return nil, nil
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
