package zttp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type App struct {
	*Router
	Routers         []*Router
	PrettyPrintJSON bool
}

// New App constructor
func NewApp() *App {
	defaultRouter := &Router{
		getRoutes:    []Route{},
		postRoutes:   []Route{},
		deleteRoutes: []Route{},
		putRoutes:    []Route{},
		patchRoutes:  []Route{},
		middlewares:  []MiddlewareWrapper{},
	}
	app := &App{
		Router:  defaultRouter,
		Routers: []*Router{defaultRouter},
	}

	defaultRouter.App = app
	return app
}

// New Router constructor
func (app *App) NewRouter(path string) *Router {
	router := &Router{
		App:          app,
		prefix:       path,
		getRoutes:    []Route{},
		postRoutes:   []Route{},
		deleteRoutes: []Route{},
		putRoutes:    []Route{},
		patchRoutes:  []Route{},
		middlewares:  []MiddlewareWrapper{},
	}

	app.Routers = append(app.Routers, router)

	return router
}

// Start listening to the given port
func (app *App) Start(port int) {

	// Initiate the tcp server sockets
	server, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println("err initiating server... " + err.Error())
	}

	for {
		// accept tcp socket connections indefinitely
		socket, err := server.Accept()
		if err != nil {
			log.Println("err accepting socket")
		}

		// handle the connected client tcp socket in a goroutine
		go handleClient(socket, app)
	}
}

// The Front Controller
// This function is responsible for handling the incoming request from the client tcp socket
// from the beginning until it sends a response and close the connection eventually
func handleClient(socket net.Conn, app *App) {
	defer socket.Close()

	// Buffer reader to read from the client tcp socket
	rdr := bufio.NewReader(socket)

	// Extract the request line, headers, and body
	requestParts := extractRequestLine(rdr, socket)
	headers, contentLength := extractHeaders(rdr)
	body := extractBody(rdr, contentLength)

	// Extract the method and the raw path from the request line
	method := requestParts[0]
	rawPath := requestParts[1]

	path := rawPath
	queries := make(map[string]string)

	// Extract queries, if exist
	if strings.Contains(rawPath, "?") {
		split := strings.SplitN(rawPath, "?", 2)
		path = split[0]
		queries = extractQueries(split[1])
	}

	// Find the matched handler from the router with parsing params, if exist
	handler, params := findHandler(method, path, socket, app)

	// If a handler matched, call it with the generated request and response objects
	// Otherwise, send a 404 not found response
	if handler != nil {
		req := Req{
			Method:  method,
			Path:    path,
			Body:    body,
			Headers: headers,
			Params:  params,
			Queries: queries,
		}
		res := Res{
			Socket:          socket,
			StatusCode:      200,
			Headers:         make(map[string]string),
			PrettyPrintJSON: app.PrettyPrintJSON,
		}

		handler(req, res)
	} else {
		sendResponse(socket, []byte("Not Found"), 404, "text/plain", nil)
	}
}
