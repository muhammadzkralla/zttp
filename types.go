package zttp

type Handler func(req Req, res Res)
type Middleware func(req Req, res Res, next func())

type Route struct {
	path    string
	handler Handler
}
