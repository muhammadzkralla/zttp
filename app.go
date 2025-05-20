package zttp

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type App struct {
	*Router
	Routers         []*Router
	PrettyPrintJSON bool
}

type Ctx struct {
	Req *Req
	Res *Res
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
		log.Fatalf("err initiating server... %s", err.Error())
	}

	defer server.Close()

	for {
		// accept tcp socket connections indefinitely
		socket, err := server.Accept()
		if err != nil {
			log.Println("err accepting socket: ", err)
			continue
		}

		// handle the connected client tcp socket in a goroutine
		go handleClient(socket, app)
	}
}

// Start listening securely to the given port
func (app *App) StartTls(port int, certFile, keyFile string) {

	// Load TLS certificate and key files
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}

	// Pass them to the TLS config
	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	// Initiate the tcp server sockets securely
	server, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), config)
	if err != nil {
		log.Fatalf("failed to start TLS listener: %s", err)
	}

	defer server.Close()

	for {
		// accept tcp socket connections indefinitely
		socket, err := server.Accept()
		if err != nil {
			log.Println("err accepting TLS connection:", err)
			continue
		}

		// handle the connected client tcp socket in a goroutine
		go handleClient(socket, app)
	}
}

// The Front Controller
// This function is responsible for handling the incoming request from the client tcp socket
// from the beginning until it sends a response and close the connection eventually
func handleClient(socket net.Conn, app *App) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			sendResponse(socket, []byte("Internal Server Error"), 500, "text/plain", nil)
		}
		socket.Close()
	}()

	// Buffer reader to read from the client tcp socket
	rdr := bufio.NewReader(socket)

	for {
		// Set hard-coded read timeout for now
		// TODO: Make it an app's config specification later
		if err := socket.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			log.Printf("Error setting read deadline: %v", err)
			return
		}

		// Extract the request line, headers, and body
		requestParts := extractRequestLine(rdr, socket)
		// TODO: make extractRequestLine() return []string, bool instead
		// NOTE: THIS WAS ADDED TO AVOID EMPTY TCP CONNECTIONS MADE BY POSTMAN
		// I think this is somehow related to the keep-alive request header
		// I will figure it out later
		if len(requestParts) < 2 {
			// Request was already handled and response sent, so just return
			return
		}

		headers, contentLength := extractHeaders(rdr)
		body := extractBody(rdr, contentLength)
		cookies := extractCookies(headers)

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
			req := &Req{
				Method:  method,
				Path:    path,
				Body:    body,
				Headers: headers,
				Params:  params,
				Queries: queries,
				Cookies: cookies,
			}
			res := &Res{
				Socket:          socket,
				StatusCode:      200,
				Headers:         make(map[string][]string),
				PrettyPrintJSON: app.PrettyPrintJSON,
			}

			ctx := &Ctx{
				Req: req,
				Res: res,
			}

			req.Ctx = ctx
			res.Ctx = ctx

			handler(req, res)
		} else {
			sendResponse(socket, []byte("Not Found"), 404, "text/plain", nil)
		}

		// Check if client requested connection close
		connectionHeader := strings.ToLower(headers["Connection"])
		if connectionHeader == "close" {
			return
		}
	}
}
