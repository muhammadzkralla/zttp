package zttp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type Req struct {
	Method  string
	Path    string
	Body    string
	Headers map[string]string
	Params  map[string]string
	Queries map[string]string
	Cookies map[string]string
}

// Return the value of the passed header key
func (req *Req) Header(key string) string {
	return req.Headers[key]
}

// Return the value of the passed param key
func (req *Req) Param(key string) string {
	return req.Params[key]
}

// Return the value of the passed query key
func (req *Req) Query(key string) string {
	return req.Queries[key]
}

// Parse the request body into the target struct
// Note that the target MUST be a pointer
func (req *Req) ParseJson(target any) error {
	return json.Unmarshal([]byte(req.Body), target)
}

// Extract the request line from the buffer of the current client tcp socket
func extractRequestLine(rdr *bufio.Reader, socket net.Conn) []string {
	var requestParts []string

	// The request line is always the first line in the request
	requestLine, err := rdr.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			log.Println("connection closed by client")
			return requestParts
		}

		log.Println("err reading from socket... " + err.Error())
		return requestParts
	}

	// Remove all leading and trailing white spaces
	requestLine = strings.TrimSpace(requestLine)

	// Log the incoming request
	// TODO: must be a configuration detail later
	fmt.Println("Incoming request: " + requestLine)

	// Request line is empty, bad request
	if requestLine == "" {
		log.Println("empty request line, sending 'Bad Request' response")
		sendResponse(socket, []byte("Bad Request"), 400, "text/plain", nil)
		return requestParts
	}

	// Split the request line into three parts and return them as a slice
	requestParts = strings.SplitN(requestLine, " ", 3)
	if len(requestParts) < 2 {
		log.Println("invalid request line: " + requestLine)
		sendResponse(socket, []byte("Bad Request"), 400, "text/plain", nil)
		return requestParts
	}

	return requestParts
}

// Extract the request headers and the body's content length (if exists) from the buffer of the current client tcp socket
func extractHeaders(rdr *bufio.Reader) (map[string]string, int) {
	headers := make(map[string]string)
	var contentLength int = 0

	// Keep reading each line and parse it as a header until reaching an empty line
	for {
		line, err := rdr.ReadString('\n')
		if err != nil {
			log.Println("err reading headers... " + err.Error())
			return nil, 0
		}

		// If the current header is the `Content-Length` header, store its value to return later
		if strings.HasPrefix(line, "Content-Length") {
			parts := strings.Split(line, ":")
			lengthStr := strings.TrimSpace(parts[1])
			contentLength, err = strconv.Atoi(lengthStr)
		}

		// Remove all leading and trailing white spaces and detect the end of the headers section
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		// Parse the header and store it
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headers[key] = value
	}

	return headers, contentLength
}

// Extract the request body from the buffer of the current client tcp socket
func extractBody(rdr *bufio.Reader, contentLength int) string {

	body := ""

	// Read exactly the next `contentLength` bytes in the buffer
	if contentLength > 0 {
		bodyBuffer := make([]byte, contentLength)
		_, err := io.ReadFull(rdr, bodyBuffer)
		if err != nil {
			log.Println("err reading body... " + err.Error())
			return ""
		}

		body = string(bodyBuffer)
	}

	return body
}

// Extract the request queries from the buffer of the current client tcp socket
func extractQueries(rawPath string) map[string]string {
	queries := make(map[string]string)

	if rawPath == "" {
		return queries
	}

	// Split the raw path with the `&` delimiter
	pairs := strings.SplitSeq(rawPath, "&")

	// Parse the query and store it
	for pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) > 1 {
			value = parts[1]
		}
		queries[key] = value
	}

	return queries
}

// Extract the request cookies from the request headers
func extractCookies(headers map[string]string) map[string]string {
	cookies := make(map[string]string)

	if cookieHeader, ok := headers["Cookie"]; ok {
		pairs := strings.Split(cookieHeader, ";")

		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			kv := strings.SplitN(pair, "=", 2)

			if len(kv) == 2 {
				cookies[kv[0]] = kv[1]
			}
		}
	}

	return cookies
}

// Find the matched handler with the passed path from the router and parse params, if exist
func findHandler(method, path string, socket net.Conn, app *App) (Handler, map[string]string) {
	for _, router := range app.Routers {
		var routes []Route
		switch method {
		case "GET":
			routes = router.getRoutes
		case "DELETE":
			routes = router.deleteRoutes
		case "POST":
			routes = router.postRoutes
		case "PUT":
			routes = router.putRoutes
		case "PATCH":
			routes = router.patchRoutes
		default:
			log.Println("unsupported method:", method)
			sendResponse(socket, []byte("Method Not Allowed"), 405, "text/plain", nil)
			return nil, nil
		}

		if handler, params := matchRoute(path, routes); handler != nil {
			return handler, params
		}
	}

	return nil, nil
}
