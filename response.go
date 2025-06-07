package zttp

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type Cookie struct {
	Name        string    `json:"name"`
	Value       string    `json:"value"`
	Path        string    `json:"path"`
	Domain      string    `json:"domain"`
	Expires     time.Time `json:"expires"`
	MaxAge      int       `json:"max_age"`
	Secure      bool      `json:"secure"`
	HttpOnly    bool      `json:"http_only"`
	SameSite    string    `json:"same_site"`
	SessionOnly bool      `json:"session_only"`
}

type Res struct {
	Socket          net.Conn
	StatusCode      int
	Headers         map[string][]string
	ContentType     string
	PrettyPrintJSON bool
	*Ctx
}

// This function sends a text/plain response body
func (res *Res) Send(data string) {
	if res.ContentType == "" {
		res.ContentType = "text/plain; charset=utf-8"
	}

	sendResponse(res.Socket, []byte(data), res.StatusCode, res.ContentType, res.Headers)
}

// This function sends a JSON response body
func (res *Res) Json(data any) {
	var raw []byte
	var err error

	// If the app is configured to pretty print JSON responses or not
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
	sendResponse(res.Socket, raw, res.StatusCode, res.ContentType, res.Headers)
}

func (res *Res) Static(path, root string) {
	// Clean the root directory path
	root = filepath.Clean(root)

	if path == "" {
		path = "/"
	}

	// Check if file exists
	fullPath := filepath.Join(root, path)
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		res.Status(404).Send("Not Found")
		return
	}

	// If it's a directory, fallback to index.html
	if fileInfo.IsDir() {
		indexPath := filepath.Join(fullPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			fullPath = indexPath
		} else {
			res.Status(403).Send("Couldn't find index.html in given directory")
			return
		}
	}

	// Open the file
	file, err := os.Open(fullPath)
	if err != nil {
		res.Status(500).Send("Internal Server Error")
		return
	}

	defer file.Close()

	// Get file info again in case we swithced to index.html
	fileInfo, err = file.Stat()
	if err != nil {
		res.Status(500).Send("Internal Server Error")
		return
	}

	// Set content type based on file extension
	ext := filepath.Ext(fullPath)
	res.ContentType = mime.TypeByExtension(ext)
	if res.ContentType == "" {
		res.ContentType = "application/octet-stream"
	}
	res.Header("Content-Type", res.ContentType)

	// Set Last-Modified header
	modTime := fileInfo.ModTime()
	res.Header("Last-Modified", modTime.UTC().Format(http.TimeFormat))

	// Handle If-Modified-Since header
	ifModifiedSince := ""
	ifModifiedSinceHeader, ok := res.Headers["If-Modified-Since"]
	if ok {
		ifModifiedSince = ifModifiedSinceHeader[0]
	}

	if ifModifiedSince != "" {
		if t, err := time.Parse(http.TimeFormat, ifModifiedSince); err == nil {
			if modTime.Before(t.Add(1 * time.Second)) {
				res.Status(304).Send("")
				return
			}
		}
	}

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		res.Status(500).Send("Internal Server Error")
		return
	}

	// Send the file content
	res.Send(string(content))
}

// This function ends the current response
func (res *Res) End() {
	if res.ContentType == "" {
		res.ContentType = "text/plain; charset=utf-8"
	}

	sendResponse(res.Socket, []byte(""), res.StatusCode, res.ContentType, res.Headers)
}

// Sets the value of the passed header key
func (res *Res) Header(key, value string) *Res {
	if _, exists := res.Headers[key]; !exists {
		res.Headers[key] = []string{}
	}

	res.Headers[key] = append(res.Headers[key], value)

	return res
}

// Sets the response cookies
func (res *Res) SetCookie(cookie Cookie) *Res {
	// Build the cookie string
	cookieStr := fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)

	// Add optional attributes
	if cookie.Path != "" {
		cookieStr += fmt.Sprintf("; Path=%s", cookie.Path)
	}
	if cookie.Domain != "" {
		cookieStr += fmt.Sprintf("; Domain=%s", cookie.Domain)
	}
	if !cookie.Expires.IsZero() {
		cookieStr += fmt.Sprintf("; Expires=%s", cookie.Expires.UTC().Format(time.RFC1123))
	}
	if cookie.MaxAge > 0 {
		cookieStr += fmt.Sprintf("; Max-Age=%d", cookie.MaxAge)
	}
	if cookie.Secure {
		cookieStr += "; Secure"
	}
	if cookie.HttpOnly {
		cookieStr += "; HttpOnly"
	}
	if cookie.SameSite != "" {
		switch cookie.SameSite {
		case "Strict", "Lax", "None":
			cookieStr += fmt.Sprintf("; SameSite=%s", cookie.SameSite)
		default:
			log.Printf("Warning: Invalid SameSite value: %s", cookie.SameSite)
		}
	}
	if cookie.SessionOnly {
		cookieStr += "; SessionOnly=true"
	}

	// Append to headers
	res.Header("Set-Cookie", cookieStr)

	return res
}

// Sets the status code of the current response
func (res *Res) Status(code int) *Res {
	res.StatusCode = code
	return res
}

// Sets the `Vary` HTTP response header
func (res *Res) Vary(fields ...string) {
	if len(fields) == 0 {
		return
	}

	_, ok := res.Headers["Vary"]
	if !ok {
		res.Header("Vary", "")
	}

	// Either will be `` or `value` or `value_1, value_2, ..., value_n`
	varyHeader := res.Headers["Vary"][0]
	varyHeaderParts := strings.Split(varyHeader, ", ")

	for i := range fields {
		fields[i] = textproto.CanonicalMIMEHeaderKey(fields[i])
	}

	for i := range varyHeaderParts {
		varyHeaderParts[i] = textproto.CanonicalMIMEHeaderKey(varyHeaderParts[i])
	}

	var newFields []string
	for _, field := range fields {
		if !containString(varyHeaderParts, field) {
			newFields = append(newFields, field)
		}
	}
	fields = newFields

	fields = removeDuplicates(fields)
	joinedFields := strings.Join(fields, ", ")

	if varyHeader == "" {
		res.Headers["Vary"][0] = joinedFields
	} else if joinedFields != "" {
		res.Headers["Vary"][0] += ", " + joinedFields
	}

	return
}

// Check if a certain string exists in a slice of strings
func containString(slice []string, target string) bool {
	return slices.Contains(slice, target)
}

// Remove duplicates from a slice of strings
func removeDuplicates(list []string) []string {
	var result []string
	visited := make(map[string]bool)

	for _, val := range list {
		if !visited[val] {
			visited[val] = true
			result = append(result, val)
		}
	}

	return result
}

// Writes the response data into the client tcp socket's buffer
func sendResponse(socket net.Conn, body []byte, code int, contentType string, headers map[string][]string) {
	statusMessage := http.StatusText(code)
	fmt.Fprintf(socket, "HTTP/1.1 %d %s\r\n", code, statusMessage)
	fmt.Fprintf(socket, "Content-Length: %d\r\n", len(body))
	fmt.Fprintf(socket, "Content-Type: %s\r\n", contentType)

	// If there's any extra response headers
	if headers != nil {
		for k, values := range headers {
			for _, v := range values {
				fmt.Fprintf(socket, "%s: %s\r\n", k, v)
			}
		}
	}
	fmt.Fprintf(socket, "\r\n")

	if body == nil {
		body = []byte{}
	}

	_, err := socket.Write(body)
	if err != nil {
		log.Println("Error writing response body:", err)
	}
}
