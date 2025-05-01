package zttp

import (
	"strings"
	"testing"
)

// A template user struct for testing
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// Test sending text response
func TestSendResponse(t *testing.T) {
	// Mock a socket
	conn := &MockConn{}

	res := Res{
		Socket:     conn,
		StatusCode: 200,
	}

	res.Send("OK")

	// Expected response body
	expected := "HTTP/1.1 200 OK\r\nContent-Length: 2\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nOK"
	if string(conn.data) != expected {
		t.Errorf("Expected response %s, but got %s", expected, string(conn.data))
	}
}

// Test sending parsed JSON response
func TestSendJsonResponse(t *testing.T) {
	// Mock a socket
	conn := &MockConn{}

	data := map[string]string{
		"message": "OK",
	}

	res := Res{
		Socket:     conn,
		StatusCode: 200,
	}

	res.Json(data)

	// Expected response body
	if !strings.Contains(string(conn.data), `"message":"OK"`) {
		t.Errorf("Expected JSON response, but got %s", string(conn.data))
	}
}

// Test sending parsed JSON response
func TestJson(t *testing.T) {
	conn := &MockConn{}
	res := &Res{
		Socket:     conn,
		StatusCode: 200,
	}

	user := User{
		Name: "Zkrallah",
		Age:  21,
	}

	res.Json(user)

	output := string(conn.data)

	if !strings.Contains(output, "Content-Type: application/json") {
		t.Errorf("Expected JSON content type, but got %s", output)
	}

	if !strings.Contains(output, `"name":"Zkrallah"`) || !strings.Contains(output, `"age":21`) {
		t.Errorf("Expected JSON body, but got %s", output)
	}
}

// Test setting response headers
func TestSetResponseHeaders(t *testing.T) {
	res := Res{
		Headers: make(map[string]string),
	}

	res.Set("Header1", "header1")
	res.Set("Header1", "notheader1")
	res.Set("Header2", "header2")

	if res.Headers["Header1"] != "notheader1" || res.Headers["Header2"] != "header2" {
		t.Errorf("Error setting response headers")
	}
}

// Test setting response status code
func TestSetStatusCode(t *testing.T) {
	res := &Res{}
	res.Status(500)
	if res.StatusCode != 500 {
		t.Errorf("Error setting status code to 500")
	}

	res.Status(400)
	res.Status(301)
	if res.StatusCode != 301 {
		t.Errorf("Error setting status code to 301")
	}

	res.Status(404)
	if res.StatusCode != 404 {
		t.Errorf("Error setting status code to 404")
	}
}
