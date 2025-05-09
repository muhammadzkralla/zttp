package zttp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Cookie struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	Expires  time.Time
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite string
}

type Res struct {
	Socket          net.Conn
	StatusCode      int
	Headers         map[string][]string
	ContentType     string
	PrettyPrintJSON bool
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
	res.ContentType = getContentType(ext)
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

// Sets the value of the passed header key
func (res *Res) Header(key, value string) {
	res.Headers[key] = append(res.Headers[key], value)
}

// Sets the status code of the current response
func (res *Res) Status(code int) *Res {
	res.StatusCode = code
	return res
}

// Writes the response data into the client tcp socket's buffer
func sendResponse(socket net.Conn, body []byte, code int, contentType string, headers map[string][]string) {
	statusMessage := getHTTPStatusMessage(code)
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

	_, err := socket.Write(body)
	if err != nil {
		log.Println("Error writing response body:", err)
	}
}

// Translate the given status code into an HTTP status message
// TODO: must be extended with more codes
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

// Translate the given extension into an HTTP response Content-Type header
// TODO: must be extended with more types
func getContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".txt":
		return "text/plain; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}
