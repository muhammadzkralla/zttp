package zttp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type App struct {
	getRoutes       []Route
	postRoutes      []Route
	deleteRoutes    []Route
	putRoutes       []Route
	patchRoutes     []Route
	middlewares     []Middleware
	PrettyPrintJSON bool
}

func (app *App) Start(port int) {
	p := fmt.Sprintf(":%d", port)
	server, err := net.Listen("tcp", p)
	if err != nil {
		log.Println("err initiating server... " + err.Error())
	}

	for {
		socket, err := server.Accept()
		if err != nil {
			log.Println("err accepting socket")
		}

		go handleClient(socket, app)
	}
}

func handleClient(socket net.Conn, app *App) {
	defer socket.Close()

	rdr := bufio.NewReader(socket)
	requestLine, err := rdr.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			log.Println("connection closed by client")
			return
		}

		log.Println("err reading from socket... " + err.Error())
		return
	}

	requestLine = strings.TrimSpace(requestLine)
	fmt.Println("Incoming request: " + requestLine)

	if requestLine == "" {
		log.Println("empty request line, sending 'Bad Request' response")
		sendResponse(socket, "Bad Request", 400, "text/plain", nil)
		return
	}

	requestParts := strings.SplitN(requestLine, " ", 3)
	if len(requestParts) < 2 {
		log.Println("invalid request line: " + requestLine)
		sendResponse(socket, "Bad Request", 400, "text/plain", nil)
		return
	}

	headers, contentLength := extractHeaders(rdr)

	body := extractBody(rdr, contentLength)

	method := requestParts[0]
	rawPath := requestParts[1]

	path := rawPath
	queries := make(map[string]string)

	if strings.Contains(rawPath, "?") {
		split := strings.SplitN(rawPath, "?", 2)
		path = split[0]
		queries = extractQueries(split[1])
	}

	var handler Handler
	var params map[string]string

	switch method {
	case "GET":
		handler, params = matchRoute(path, app.getRoutes)
	case "DELETE":
		handler, params = matchRoute(path, app.deleteRoutes)
	case "POST":
		handler, params = matchRoute(path, app.postRoutes)
	case "PUT":
		handler, params = matchRoute(path, app.putRoutes)
	case "PATCH":
		handler, params = matchRoute(path, app.patchRoutes)
	default:
		log.Println("unsupported method:", method)
		sendResponse(socket, "Method Not Allowed", 405, "text/plain", nil)
		return
	}

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
			StatusCode:          200,
			Headers:         make(map[string]string),
			PrettyPrintJSON: app.PrettyPrintJSON,
		}

		handler(req, res)
	} else {
		sendResponse(socket, "Not Found", 404, "text/plain", nil)
	}
}
