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

// A template user struct for testing
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

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

// Test DELETE route matching
func TestDeleteRouteMatching(t *testing.T) {
	app := &App{}

	// Mock a DELETE handler
	app.Delete("/test", func(req Req, res Res) {
		res.Send("DELETE route matched")
	})

	// Mock a DELETE request
	response := mockRequest(app, "DELETE", "/test", "")

	if !strings.Contains(response, "DELETE route matched") {
		t.Errorf("Expected response to contain 'DELETE route matched', but got %s", response)
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

// Test PUT route matching
func TestPutRouteMatching(t *testing.T) {
	app := &App{}

	// Mock a PUT handler
	app.Put("/test", func(req Req, res Res) {
		res.Send("PUT route matched")
	})

	// Mock a PUT request
	response := mockRequest(app, "PUT", "/test", "")

	if !strings.Contains(response, "PUT route matched") {
		t.Errorf("Expected response to contain 'PUT route matched', but got %s", response)
	}
}

// Test PATCH route matching
func TestPatchRouteMatching(t *testing.T) {
	app := &App{}

	// Mock a PATCH handler
	app.Patch("/test", func(req Req, res Res) {
		res.Send("PATCH route matched")
	})

	// Mock a PATCH request
	response := mockRequest(app, "PATCH", "/test", "")

	if !strings.Contains(response, "PATCH route matched") {
		t.Errorf("Expected response to contain 'PATCH route matched', but got %s", response)
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
	expected := "HTTP/1.1 200 OK\r\nContent-Length: 2\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nOK"
	if string(conn.data) != expected {
		t.Errorf("Expected response %s, but got %s", expected, string(conn.data))
	}
}

func TestSendJsonResponse(t *testing.T) {
	// Mock a socket
	conn := &MockConn{}

	data := map[string]string{
		"message": "OK",
	}

	res := Res{
		Socket: conn,
		Status: 200,
	}

	res.Json(data)

	// Expected response body
	if !strings.Contains(string(conn.data), `"message":"OK"`) {
		t.Errorf("Expected JSON response, but got %s", string(conn.data))
	}
}

func TestExtractHeader(t *testing.T) {
	headers := map[string]string{
		"Content-Type":   "application/json",
		"Content-Length": "20",
		"Header1":        "header1",
		"Header2":        "header2",
	}

	req := Req{
		Headers: headers,
	}

	h1 := req.Header("Content-Type")
	h2 := req.Header("Content-Length")
	h3 := req.Header("Header1")
	h4 := req.Header("Header2")
	h5 := req.Header("unknown")

	if h1 != "application/json" || h2 != "20" || h3 != "header1" || h4 != "header2" || h5 != "" {
		t.Errorf("Error parsing headers")
	}
}

func TestExtractHeaders(t *testing.T) {
	// Mock some headers and create a bufio reader of them
	headers := "Content-Length: 20\r\nHeader1: header1\r\nHeader2: header2\r\n\r\n"
	rdr := bufio.NewReader(bytes.NewBufferString(headers))

	extractedHeaders, extractedLen := extractHeaders(rdr)

	if extractedHeaders["Content-Length"] != "20" {
		t.Errorf("Expected header 'Content-Length: 20', but got %s", extractedHeaders["Content-Length"])
	}

	if extractedHeaders["Header1"] != "header1" {
		t.Errorf("Expected header 'Header1: header1', but got %s", extractedHeaders["Header1"])
	}

	if extractedHeaders["Header2"] != "header2" {
		t.Errorf("Expected header 'Header2: header2', but got %s", extractedHeaders["Header2"])
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

func TestJson(t *testing.T) {
	socket := &MockConn{}
	res := &Res{
		Socket: socket,
		Status: 200,
	}

	user := User{
		Name: "Zkrallah",
		Age:  21,
	}

	res.Json(user)

	output := string(socket.data)

	if !strings.Contains(output, "Content-Type: application/json") {
		t.Errorf("Expected JSON content type, but got %s", output)
	}

	if !strings.Contains(output, `"name":"Zkrallah"`) || !strings.Contains(output, `"age":21`) {
		t.Errorf("Expected JSON body, but got %s", output)
	}
}

func TestParseJson(t *testing.T) {
	// Mock an incoming json request
	body := `{"name":"Zkrallah","age":21}`
	req := Req{
		Body: body,
	}

	// attempt parsing the request json into the user struct object
	var user User
	err := req.ParseJson(&user)
	if err != nil {
		t.Errorf("Expected no error while parsing json, got %v", err)
	}

	if user.Name != "Zkrallah" || user.Age != 21 {
		t.Errorf("Expected no error while parsing json, got %v", err)
	}
}

func TestParseQueries(t *testing.T) {
	q1 := "userId=2&name=zkrallah&category=admin"
	q2 := "userId=1&category=teacher&limit=2"
	q3 := "userId=1&category="
	q4 := ""

	qs1 := extractQueries(q1)
	qs2 := extractQueries(q2)
	qs3 := extractQueries(q3)
	qs4 := extractQueries(q4)

	if len(qs1) != 3 || len(qs2) != 3 || len(qs3) != 2 || len(qs4) != 0 {
		t.Errorf("Sizes are not correct: %d", len(qs4))
	}

	if qs1["userId"] != "2" || qs1["name"] != "zkrallah" || qs1["category"] != "admin" {
		t.Errorf("Error in parsing first query")
	}

	if qs2["userId"] != "1" || qs2["category"] != "teacher" || qs2["limit"] != "2" {
		t.Errorf("Error in parsing second query")
	}

	if qs3["userId"] != "1" || qs3["category"] != "" {
		t.Errorf("Error in parsing third query")
	}
}

func TestParseQuery(t *testing.T) {
	queries := map[string]string{
		"userId":   "2",
		"name":     "zkrallah",
		"category": "admin",
	}

	req := Req{
		Queries: queries,
	}

	q1 := req.Query("userId")
	q2 := req.Query("name")
	q3 := req.Query("category")
	q4 := req.Query("unknown")

	if q1 != "2" || q2 != "zkrallah" || q3 != "admin" || q4 != "" {
		t.Errorf("Error parsing queries")
	}
}

func TestSetResponseHeaders(t *testing.T) {
	conn := &MockConn{}

	res := Res{
		Socket:  conn,
		Headers: make(map[string]string),
	}

	res.Set("Header1", "header1")
	res.Set("Header1", "notheader1")
	res.Set("Header2", "header2")

	if res.Headers["Header1"] != "notheader1" || res.Headers["Header2"] != "header2" {
		t.Errorf("Error setting response headers")
	}
}
