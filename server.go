package zttp

import (
	"bufio"
	"fmt"
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
		log.Println("err reading from socket... " + err.Error())
	}

	fmt.Println("Incoming request: " + requestLine)

	_, contentLength := extractHeaders(rdr)

	body := extractBody(rdr, contentLength)

	requestParts := strings.Split(requestLine, " ")
	method := requestParts[0]
	endPoint := requestParts[1]

	var handler Handler
	var params map[string]string

	if method == "GET" {
		handler, params = matchRoute(endPoint, app.getRoutes)
	} else if method == "DELETE" {
		handler, params = matchRoute(endPoint, app.deleteRoutes)
	} else if method == "POST" {
		handler, params = matchRoute(endPoint, app.postRoutes)
	} else if method == "PUT" {
		handler, params = matchRoute(endPoint, app.putRoutes)
	} else if method == "PATCH" {
		handler, params = matchRoute(endPoint, app.patchRoutes)
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
