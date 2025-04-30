package zttp

import "strings"

type Route struct {
	path    string
	handler Handler
}

type Handler func(req Req, res Res)

func (app *App) Get(endpoint string, handler Handler) {
	app.getRoutes = append(app.getRoutes, Route{endpoint, applyMiddleware(handler, app)})
}

func (app *App) Delete(endpoint string, handler Handler) {
	app.deleteRoutes = append(app.deleteRoutes, Route{endpoint, applyMiddleware(handler, app)})
}

func (app *App) Post(endpoint string, handler Handler) {
	app.postRoutes = append(app.postRoutes, Route{endpoint, applyMiddleware(handler, app)})
}

func (app *App) Put(endpoint string, handler Handler) {
	app.putRoutes = append(app.putRoutes, Route{endpoint, applyMiddleware(handler, app)})
}

func (app *App) Patch(endpoint string, handler Handler) {
	app.patchRoutes = append(app.patchRoutes, Route{endpoint, applyMiddleware(handler, app)})
}

func matchRoute(requestPath string, routes []Route) (Handler, map[string]string) {
	for _, route := range routes {
		params := make(map[string]string)
		routeParts := strings.Split(route.path, "/")
		requestParts := strings.Split(requestPath, "/")

		if len(routeParts) != len(requestParts) {
			continue
		}

		match := true
		for i := range routeParts {
			if strings.HasPrefix(routeParts[i], ":") {
				paramName := routeParts[i][1:]
				params[paramName] = requestParts[i]
			} else if routeParts[i] != requestParts[i] {
				match = false
				break
			}
		}

		if match {
			return route.handler, params
		}
	}

	return nil, nil
}
