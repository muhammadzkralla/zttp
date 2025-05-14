package zttp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
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
	*Ctx
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

// Return true when the response is still “fresh” in the client's cache.
// otherwise false is returned to indicate that the client cache is now stale
// and the full response should be sent.
// When a client sends the Cache-Control: no-cache request header to indicate an end-to-end
// reload request, this will return false to make handling these requests transparent.
// This logic is heavily inspired by the official gofiber source code, with some touches of mine:
// https://github.com/gofiber/fiber/blob/main/ctx.go
func (req *Req) Fresh() bool {
	etagMatched := true
	modifiedSinceMatched := true
	etagMissing := false

	// Check for conditional request headers
	modifiedSince := req.Header("If-Modified-Since")
	noneMatch := req.Header("If-None-Match")

	// The request is unconditional
	if modifiedSince == "" && noneMatch == "" {
		log.Println("The request is unconditional")
		return false
	}

	// Check `Cache-Control` request header to see if the
	// request is intended to be an end-to-end request
	cacheControl := req.Header("Cache-Control")
	if cacheControl != "" && hasNoCacheDirective(cacheControl) {
		log.Println("The request has the `Cache-Control: no-cache` header")
		return false
	}

	// Start comparing conditional request headers with response headers
	if noneMatch != "" && noneMatch != "*" {
		etags := req.Ctx.Res.Headers["ETag"]
		if len(etags) == 0 {
			log.Println("`ETag` response header not found")
			etagMatched = false
		}
		etag := etags[0]

		if etag == "" {
			log.Println("`ETag` response header not found")
			etagMatched = false
		}

		// Check `Etag` and `If-None-Match` headers first
		if isEtagStale(etag, []byte(noneMatch)) {
			log.Println("`ETAG` response header didn't match with `If-None-Match` request header")
			etagMatched = false
		}
	} else {
		etagMatched = false
		etagMissing = true
	}

	if modifiedSince != "" {
		lastModifiedHeaders := req.Ctx.Res.Headers["Last-Modified"]
		if len(lastModifiedHeaders) == 0 {
			log.Println("`Last-Modified` response header not found")
			modifiedSinceMatched = false
		}

		lastModified := lastModifiedHeaders[0]

		if lastModified == "" {
			log.Println("`Last-Modified` response header not found")
			modifiedSinceMatched = false
		}

		if lastModified != "" {
			lastModifiedTime, err := http.ParseTime(lastModified)
			if err != nil {
				log.Println("Could not parse last modified time")
				modifiedSinceMatched = false
			}

			modifiedSinceTime, err := http.ParseTime(modifiedSince)
			if err != nil {
				log.Println("Could not parse modified since time")
				modifiedSinceMatched = false
			}

			// Return true if modifiedSinceTime is not after lastModifiedTime
			if lastModifiedTime.After(modifiedSinceTime) {
				log.Println("Resource modified")
				modifiedSinceMatched = false
			} else {
				log.Println("Resource wasn't modified")
			}
		}
	}

	return etagMatched || (etagMissing && modifiedSinceMatched)
}

// If the request is not fresh, then it's stale
func (req *Req) Stale() bool {
	return !req.Fresh()
}

// Check if the Cache-Control header contains a valid 'no-cache' directive
func hasNoCacheDirective(cacheControl string) bool {
	const directive = "no-cache"

	// Check if the directive exists at all
	pos := strings.Index(cacheControl, directive)
	if pos < 0 {
		return false
	}

	// Check the character before the directive
	if pos > 0 {
		prevChar := cacheControl[pos-1]
		if prevChar != ' ' && prevChar != ',' {
			return false
		}
	}

	endPos := pos + len(directive)

	// Case 1: Directive at end of string
	if endPos == len(cacheControl) {
		return true
	}

	// Case 2: Check character after directive
	nextChar := cacheControl[endPos]
	return nextChar == ',' || nextChar == ' '
}

// Checks if two ETags match according to RFC 7232
func compareETags(clientTag, serverTag string) bool {
	// Direct match
	if clientTag == serverTag {
		return true
	}

	// Weak tag comparison cases
	if strings.HasPrefix(clientTag, "W/") {
		return clientTag[2:] == serverTag || "W/"+serverTag == clientTag
	}
	if strings.HasPrefix(serverTag, "W/") {
		return serverTag[2:] == clientTag || "W/"+clientTag == serverTag
	}

	return false
}

// Directly taken from the official gofiber source code:
// https://github.com/gofiber/fiber/blob/main/ctx.go
func isEtagStale(etag string, noneMatchBytes []byte) bool {
	var start, end int

	// Adapted from:
	// https://github.com/jshttp/fresh/blob/10e0471669dbbfbfd8de65bc6efac2ddd0bfa057/index.js#L110
	for i := range noneMatchBytes {
		switch noneMatchBytes[i] {
		case 0x20:
			if start == end {
				start = i + 1
				end = i + 1
			}
		case 0x2c:
			if compareETags(string(noneMatchBytes[start:end]), etag) {
				return false
			}
			start = i + 1
			end = i + 1
		default:
			end = i + 1
		}
	}

	return !compareETags(string(noneMatchBytes[start:end]), etag)
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
			kv := strings.Split(pair, "=")

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
