package zttp

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// Mock of net.Conn struct following the net.Conn interface specifications
type MockConn struct {
	data []byte
}

// All these functions are mocked to follow the net.Conn interface specifications
// for testing purposes only
func (m *MockConn) Read(p []byte) (n int, err error) {
	copy(p, m.data)
	return len(m.data), nil
}

func (m *MockConn) Write(p []byte) (n int, err error) {
	m.data = append(m.data, p...)
	return len(p), nil
}

func (m *MockConn) Close() error {
	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return nil
}

func (m *MockConn) RemoteAddr() net.Addr {
	return nil
}

func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// Helper function to mock a request and return the response
func mockRequest(app *App, method, path, body string) string {
	conn := &MockConn{}
	req := fmt.Sprintf("%s %s HTTP/1.1\r\nContent-Length: %d\r\n\r\n%s", method, path, len(body), body)
	conn.data = []byte(req)

	// Call handleClient with the mocked connection
	handleClient(conn, app)

	return string(conn.data)
}

// Test GET route matching
func TestGetRouteMatching(t *testing.T) {
	app := &App{}

	// Mock a GET handler
	app.Get("/test", func(req Req, res Res) {
		res.Send("GET route matched")
	})

	// Mock a GET request
	response := mockRequest(app, "GET", "/test", "")

	if !strings.Contains(response, "GET route matched") {
		t.Errorf("Expected response to contain 'GET route matched', but got %s", response)
	}
}

// Test POST route matching
func TestPostRouteMatching(t *testing.T) {
	app := &App{}

	// Mock a POST handler
	app.Post("/test", func(req Req, res Res) {
		res.Send("POST route matched")
	})

	// Mock a POST request
	response := mockRequest(app, "POST", "/test", "")

	if !strings.Contains(response, "POST route matched") {
		t.Errorf("Expected response to contain 'POST route matched', but got %s", response)
	}
}

// Test middleware
func TestMiddleware(t *testing.T) {
	app := &App{}

	// Mock a middleware
	app.Use(func(req Req, res Res, next func()) {
		res.Send("Middleware worked")
		next()
	})

	// Mock a handler
	app.Get("/test", func(req Req, res Res) {

	})

	response := mockRequest(app, "GET", "/test", "")

	if !strings.Contains(response, "Middleware worked") {
		t.Errorf("Expected response to contain 'Middleware worked', but got %s", response)
	}
}

// Test dynamic routing
func TestDynamicRouting(t *testing.T) {
	app := &App{}

	// Mock a GET handler
	app.Get("/test/:postId/comment/:commentId", func(req Req, res Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		res.Send("Post ID: " + postId + ", Comment ID: " + commentId)
	})

	// Mock a POST handler
	app.Post("/test/:postId/comment/:commentId", func(req Req, res Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		res.Send("Post ID: " + postId + ", Comment ID: " + commentId)
	})

	getResponse := mockRequest(app, "GET", "/test/123/comment/comment1", "")
	postResponse := mockRequest(app, "POST", "/test/123/comment/comment1", "")

	if !strings.Contains(getResponse, "Post ID: 123, Comment ID: comment1") {
		t.Errorf("Expected response to contain 'Post ID:123, Comment ID:comment1', but got %s", getResponse)
	}

	if !strings.Contains(postResponse, "Post ID: 123, Comment ID: comment1") {
		t.Errorf("Expected response to contain 'Post ID:123, Comment ID:comment1', but got %s", postResponse)
	}
}

func TestSendResponse(t *testing.T) {
	// Mock a socket
	conn := &MockConn{}

	res := Res{
		Socket: conn,
		Status: 200,
	}

	res.Send("OK")

	// Expected response body
	expected := "HTTP/1.1 200 OK\r\nContent-Length: 2\r\nContent-Type: text/plain\r\n\r\nOK"
	if string(conn.data) != expected {
		t.Errorf("Expected response %s, but got %s", expected, string(conn.data))
	}
}

func TestExtractHeaders(t *testing.T) {
	// Mock some headers and create a bufio reader of them
	headers := "Content-Length: 20\r\nHeader1: header1\r\nHeader2: header2\r\n\r\n"
	rdr := bufio.NewReader(bytes.NewBufferString(headers))

	extractedHeaders, extractedLen := extractHeaders(rdr)

	if extractedHeaders[0] != "Content-Length: 20" {
		t.Errorf("Expected header 'Content-Length: 20', but got %s", extractedHeaders[0])
	}

	if extractedHeaders[1] != "Header1: header1" {
		t.Errorf("Expected header 'Header1: header1', but got %s", extractedHeaders[1])
	}

	if extractedHeaders[2] != "Header2: header2" {
		t.Errorf("Expected header 'Header2: header2', but got %s", extractedHeaders[2])
	}

	if extractedLen != 20 {
		t.Errorf("Expected Content-Length 20, but got %d", extractedLen)
	}
}

func TestExtractBody(t *testing.T) {
	// Mock a request body
	body := "Hello, world!"
	rdr := bufio.NewReader(bytes.NewBufferString(body))

	extractedBody := extractBody(rdr, 13)

	if extractedBody != "Hello, world!" {
		t.Errorf("Expected body to be 'Hello, world!', but got %s", extractedBody)
	}
}

func TestNotFoundHandler(t *testing.T) {
	app := &App{}

	// Perform a request to a non-existing handler
	response := mockRequest(app, "GET", "/test", "")

	if !strings.Contains(response, "Not Found") {
		t.Errorf("Expected 'Not Found', but got %s", response)
	}
}
