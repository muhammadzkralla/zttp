package zttp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type Res struct {
	Socket          net.Conn
	Status          int
	PrettyPrintJSON bool
}

func (res *Res) Send(data string) {
	sendResponse(res.Socket, data, res.Status)
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
		res.Send("Internal Server Error: JSON Marshal Failed")
		return
	}

	sendResponse(res.Socket, string(raw), res.Status)
}

func sendResponse(socket net.Conn, body string, code int) {
	statusMessage := getHTTPStatusMessage(code)
	fmt.Fprintf(socket, "HTTP/1.1 %d %s\r\n", code, statusMessage)
	fmt.Fprintf(socket, "Content-Length: %d\r\n", len(body))
	fmt.Fprintf(socket, "Content-Type: application/json\r\n")
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
