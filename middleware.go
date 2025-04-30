package zttp

type Middleware func(req Req, res Res, next func())

func (app *App) Use(m Middleware) {
	app.middlewares = append(app.middlewares, m)
}

func applyMiddleware(finalHandler Handler, app *App) Handler {
	return func(req Req, res Res) {
		i := 0

		var next func()
		next = func() {
			if i < len(app.middlewares) {
				middleware := app.middlewares[i]
				i++
				middleware(req, res, next)
			} else {
				finalHandler(req, res)
			}
		}

		next()
	}
}
