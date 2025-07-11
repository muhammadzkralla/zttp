package zttp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	// drwxr-xr-x
	DefaultDirPerm = os.ModeDir | 0755
	// -rw-------
	DefaultFilePerm = 0600
)

type AcceptPart struct {
	part string
	q    float32
}

type FormFile struct {
	Filename string
	Content  []byte
	Header   textproto.MIMEHeader
}

type Req struct {
	LocalAddress string
	Method       string
	Path         string
	Body         string
	Headers      map[string]string
	Params       map[string]string
	Queries      map[string]string
	Cookies      map[string]string
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

// Return the reference to the app this request is associated with
func (req *Req) App() *App {
	return req.App()
}

// Return the base URL of the request derived from the `Host` HTTP header
// TODO: Should check the `X-Forwarded-Host` HTTP header also
func (req *Req) Host() string {
	return req.Header("Host")
}

// Return the value of the specified part if the request is multipart
func (req *Req) FormValue(key string) string {

	// Get the multipart reader if the request is multipart
	form, err := parseMultipart(req.Headers, []byte(req.Body))
	if err != nil {
		return ""
	}

	defer form.RemoveAll()

	// Return the first matching part value
	values := form.Value[key]
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// Return the file of the specified part if the request is multipart
func (req *Req) FormFile(name string) (*FormFile, error) {

	// Get the multipart reader if the request is multipart
	form, err := parseMultipart(req.Headers, []byte(req.Body))
	if err != nil {
		return nil, err
	}

	defer form.RemoveAll()

	// Return the first matching part value
	files := form.File[name]
	if len(files) == 0 {
		return nil, fmt.Errorf("file %s not found", name)
	}

	// Open the file to prepare the bytes we will store in the content field of the
	// FormFile struct
	fileHeader := files[0]
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}

	defer file.Close()

	// Copy the bytes
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return &FormFile{
		Filename: fileHeader.Filename,
		Content:  content,
		Header:   fileHeader.Header,
	}, nil
}

// Save the multipart form file directly to disk
// TODO: I think permissions should be a config later
func (req *Req) Save(formFile *FormFile, destination string) error {

	// Check if formFile is nil first
	if formFile == nil {
		return fmt.Errorf("nil FormFile")
	}

	// Join the file name with the specified destination
	// TODO: Should be sanitized first
	fullPath := filepath.Join(destination, formFile.Filename)

	// Create the directory
	err := os.MkdirAll(destination, DefaultDirPerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file to the destination
	err = os.WriteFile(fullPath, formFile.Content, DefaultFilePerm)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	// No errors happened
	return nil
}

// Checks if the specified types are accepted from the HTTP client
// TODO: Fix Canonicalization
func (req *Req) Accepts(offered ...string) string {
	acceptHeader := req.Header("Accept")

	if acceptHeader == "" || len(offered) == 0 {
		// TODO: Align with RFC 9110 standards
		if len(offered) != 0 {
			return offered[0]
		} else {
			return ""
		}
	}

	clientTypes := parseAcceptHeader(acceptHeader)

	for _, clientType := range clientTypes {
		for _, offered := range offered {
			if matches(clientType.part, offered) {
				return offered
			}
		}
	}

	return ""
}

// Checks if the specified types are accepted from the HTTP client
func (req *Req) AcceptsCharsets(offered ...string) string {
	charsetHeader := req.Header("Accept-Charset")
	if charsetHeader == "" {
		// TODO: Align with RFC 2616 standards
		if len(offered) != 0 {
			return offered[0]
		} else {
			return ""
		}
	}

	clientCharsets := parseAcceptHeader(charsetHeader)

	// Handle wildcard
	for _, cc := range clientCharsets {
		if cc.part == "*" && cc.q > 0 {
			return offered[0]
		}
	}

	for _, cc := range clientCharsets {
		for _, charset := range offered {
			if strings.EqualFold(cc.part, charset) && cc.q > 0 {
				return charset
			}
		}
	}

	return ""
}

func (req *Req) AcceptsEncodings(offered ...string) string {
	encodingsHeader := req.Header("Accept-Encoding")
	if encodingsHeader == "" {
		// TODO: Align with RFC 9110 standards
		if len(offered) != 0 {
			return offered[0]
		} else {
			return ""
		}
	}

	clientEncodings := parseAcceptHeader(encodingsHeader)

	// Special cases (RFC 7231)
	for _, enc := range clientEncodings {
		// "identity" is always acceptable unless explicitly forbidden with q=0
		if enc.part == "identity" && enc.q == 0 {
			// Client explicitly refuses identity
			return ""
		}
	}

	for _, enc := range clientEncodings {
		for _, offeredEnc := range offered {
			if strings.EqualFold(enc.part, offeredEnc) && enc.q > 0 {
				return offeredEnc
			}
		}
	}

	// Check for wildcard
	for _, enc := range clientEncodings {
		if enc.part == "*" && enc.q > 0 {
			return offered[0]
		}
	}

	return ""
}

func (req *Req) AcceptsLanguages(offered ...string) string {
	langHeader := req.Header("Accept-Language")
	if langHeader == "" {
		// TODO: Align with RFC 9110 standards
		if len(offered) != 0 {
			return offered[0]
		} else {
			return ""
		}
	}

	acceptedLangs := parseAcceptHeader(langHeader)

	// Check for wildcard
	for _, lang := range acceptedLangs {
		if lang.part == "*" && lang.q > 0 {
			return offered[0]
		}
	}

	// Check exact match
	for _, lang := range acceptedLangs {
		for _, offeredLang := range offered {
			if strings.EqualFold(lang.part, offeredLang) && lang.q > 0 {
				return offeredLang
			}
		}
	}

	// Check primary language matches like `en` matches `en-US`
	for _, lang := range acceptedLangs {
		if lang.q == 0 {
			continue
		}

		primaryLang := strings.Split(lang.part, "-")[0]
		for _, offeredLang := range offered {
			offeredPrimary := strings.Split(offeredLang, "-")[0]
			if strings.EqualFold(primaryLang, offeredPrimary) {
				return offeredLang
			}
		}
	}

	return ""
}

// TODO: Align with RFC 7239 standards
func (req *Req) IP() string {
	// First, try the `X-Forwarded-For` request header
	xff := req.Header("X-Forwarded-For")
	if xff != "" {
		// Could be a list: client, proxy1, proxy2
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		host, _, err := net.SplitHostPort(ip)
		if err != nil {
			return ip
		}

		return host
	}

	// Second, try the `X-Real-IP` request header
	realIP := req.Header("X-Real-IP")
	if realIP != "" {
		ip := strings.TrimSpace(realIP)
		host, _, err := net.SplitHostPort(ip)
		if err != nil {
			return ip
		}

		return host
	}

	// Fallback to the default request tcp socket local address
	host, _, err := net.SplitHostPort(req.LocalAddress)
	if err != nil {
		// If an error happened during the split, just return it with the port number
		return req.LocalAddress
	}

	return host
}

// TODO: Align with RFC 7230 standards
// TODO: Handle IPv4 and IPv6 hosts
func (req *Req) Hostname() string {
	host := req.Header("Host")
	if host == "" {
		return ""
	}

	// Some gophery ahhh type code
	if parsedHost, port, err := net.SplitHostPort(host); err == nil {
		if _, err := strconv.Atoi(port); err != nil {
			return ""
		}
		return parsedHost
	}

	// Check if it's just an IPv6 address without port
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return host[1 : len(host)-1]
	}

	// Check if it's just an invalid IPv6 address without port
	if strings.HasPrefix(host, "[") != strings.HasSuffix(host, "]") {
		return ""
	}

	// Check for multiple colons in a non-IPv6 format
	// I'm not really comfortable with this check
	// but it passes the current tests, so I just keep it
	if strings.Count(host, ":") > 1 {
		return ""
	}

	// Guard against obvious invalid characters
	if strings.ContainsAny(host, "/\\@") {
		return ""
	}

	return strings.ToLower(strings.TrimSpace(host))
}

// Parse the `Accepts` HTTP client header and return a list of accepted types with their quality factors
func parseAcceptHeader(header string) []AcceptPart {
	var items []AcceptPart
	parts := strings.Split(header, ",")

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		segments := strings.Split(trimmed, ";")

		mime := segments[0]
		// Default q-value
		q := float32(1.0)

		// Parse quality factor if present
		if len(segments) > 1 && strings.HasPrefix(segments[1], "q=") {
			fmt.Sscanf(segments[1][2:], "%f", &q)
		}

		items = append(items, AcceptPart{part: mime, q: q})
	}

	// Sort the accepted types according to the quality factor from highest to lowest
	sort.Slice(items, func(i, j int) bool {
		return items[i].q > items[j].q
	})

	return items
}

// Checks for matching between the types of the client header and the specified type
// TODO: Support case-insensitivity later
func matches(clientType, offeredType string) bool {
	// Exact match
	if clientType == offeredType {
		return true
	}

	// Wildcard support
	if strings.HasSuffix(clientType, "/*") && clientType != "*/*" {
		return strings.Split(clientType, "/")[0] == strings.Split(offeredType, "/")[0]
	}

	// Catch-all wildcard
	return clientType == "*/*"
}

// Checks if the request is multipart or not and return back the multipart form reader reference
func parseMultipart(headers map[string]string, body []byte) (*multipart.Form, error) {
	// Check if it's a multipart request or not
	contentType := headers["Content-Type"]
	if contentType == "" {
		return nil, http.ErrNotMultipart
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, http.ErrNotMultipart
	}

	// Extract the boundary that separates between different parts
	boundary := params["boundary"]
	if boundary == "" {
		log.Println("no boundary found in Content-Type")
		return nil, http.ErrNotMultipart
	}

	// Create a multipart reader from the request body and the parsed boundary
	reader := multipart.NewReader(bytes.NewReader(body), boundary)

	// 32 MB memory limit + 10 MB added by default
	// TODO: Should be a configuration later
	return reader.ReadForm(32 << 20)
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
