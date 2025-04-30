package zttp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type Res struct {
	Socket          net.Conn
	StatusCode      int
	Headers         map[string]string
	ContentType     string
	PrettyPrintJSON bool
}

func (res *Res) Send(data string) {
	if res.ContentType == "" {
		res.ContentType = "text/plain; charset=utf-8"
	}

	sendResponse(res.Socket, data, res.StatusCode, res.ContentType, res.Headers)
}

func (res *Res) Json(data any) {
	var raw []byte
	var err error

	if res.PrettyPrintJSON {
		raw, err = json.MarshalIndent(data, "", "    ")
	} else {
		raw, err = json.Marshal(data)
	}

	if err != nil {
		log.Println("Error parsing json")
		res.StatusCode = 500
		res.Send("Internal Server Error: JSON Marshal Failed")
		return
	}

	res.ContentType = "application/json"
	sendResponse(res.Socket, string(raw), res.StatusCode, res.ContentType, res.Headers)
}

func (res *Res) Set(key, value string) {
	res.Headers[key] = value
}

func (res *Res) Status(code int) *Res {
	res.StatusCode = code
	return res
}

func sendResponse(socket net.Conn, body string, code int, contentType string, headers map[string]string) {
	statusMessage := getHTTPStatusMessage(code)
	fmt.Fprintf(socket, "HTTP/1.1 %d %s\r\n", code, statusMessage)
	fmt.Fprintf(socket, "Content-Length: %d\r\n", len(body))
	fmt.Fprintf(socket, "Content-Type: %s\r\n", contentType)
	if headers != nil {
		for k, v := range headers {
			fmt.Fprintf(socket, "%s: %s\r\n", k, v)
		}
	}
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
