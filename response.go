package zttp

import (
	"fmt"
	"net"
)

type Res struct {
	Socket net.Conn
	Status int
}

func (res *Res) Send(data string) {
	sendResponse(res.Socket, data, res.Status)
}

func sendResponse(socket net.Conn, body string, code int) {
	statusMessage := getHTTPStatusMessage(code)
	fmt.Fprintf(socket, "HTTP/1.1 %d %s\r\n", code, statusMessage)
	fmt.Fprintf(socket, "Content-Length: %d\r\n", len(body))
	fmt.Fprintf(socket, "Content-Type: text/plain\r\n")
	fmt.Fprintf(socket, "\r\n")
	fmt.Fprintf(socket, "%s", body)
}

func getHTTPStatusMessage(code int) string {
	statusMessages := map[int]string{
		200: "OK",
		201: "Created",
		400: "Bad Request",
		404: "Not Found",
		500: "Internal Server Error",
	}
	if msg, exists := statusMessages[code]; exists {
		return msg
	}
	return "Unknown Status"
}
