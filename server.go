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
	getRoutes    []Route
	postRoutes   []Route
	deleteRoutes []Route
	putRoutes    []Route
	patchRoutes  []Route
	middlewares  []Middleware
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
		sendResponse(socket, "Bad Request", 400)
		return
	}

	requestParts := strings.SplitN(requestLine, " ", 3)
	if len(requestParts) < 2 {
		log.Println("invalid request line: " + requestLine)
		sendResponse(socket, "Bad Request", 400)
		return
	}

	_, contentLength := extractHeaders(rdr)

	body := extractBody(rdr, contentLength)

	method := requestParts[0]
	endPoint := requestParts[1]

	var handler Handler
	var params map[string]string

	switch method {
	case "GET":
		handler, params = matchRoute(endPoint, app.getRoutes)
	case "DELETE":
		handler, params = matchRoute(endPoint, app.deleteRoutes)
	case "POST":
		handler, params = matchRoute(endPoint, app.postRoutes)
	case "PUT":
		handler, params = matchRoute(endPoint, app.putRoutes)
	case "PATCH":
		handler, params = matchRoute(endPoint, app.patchRoutes)
	default:
		log.Println("unsupported method:", method)
		sendResponse(socket, "Method Not Allowed", 405)
		return
	}

	if handler != nil {
		req := Req{
			Method: method,
			Path:   endPoint,
			Body:   body,
			Params: params,
		}
		res := Res{
			Socket: socket,
			Status: 200,
		}

		handler(req, res)
	} else {
		sendResponse(socket, "Not Found", 404)
	}
}
