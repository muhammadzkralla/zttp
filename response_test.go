package zttp

import (
	"strings"
	"testing"
	"time"
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
	if string(conn.outBuf) != expected {
		t.Errorf("Expected response %s, but got %s", expected, string(conn.outBuf))
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
	if !strings.Contains(string(conn.outBuf), `"message":"OK"`) {
		t.Errorf("Expected JSON response, but got %s", string(conn.outBuf))
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

	output := string(conn.outBuf)

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
		Headers: make(map[string][]string),
	}

	res.Header("Header1", "header1")
	res.Header("Header1", "notheader1")
	res.Header("Header2", "header2")

	if res.Headers["Header1"][0] != "header1" || res.Headers["Header1"][1] != "notheader1" || res.Headers["Header2"][0] != "header2" {
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

func TestStaticResponseServing(t *testing.T) {
	conn := &MockConn{}
	res := &Res{
		Socket:  conn,
		Headers: make(map[string][]string),
	}

	res.Static("index.html", "./examples/static-file-serving/public/")

	output := string(conn.outBuf)

	if res.Headers["Content-Type"][0] != "text/html; charset=utf-8" {
		t.Errorf("Expected header Content-Type: text/html; charset=utf-8, but got %s", res.Headers["Content-Type"][0])
	}

	if !strings.Contains(output, "<h1>Hello from static index file!</h1>") {
		t.Errorf("Unexpected response body: %s", output)
	}

	res = &Res{
		Socket:  conn,
		Headers: make(map[string][]string),
	}

	res.Static("home.html", "./examples/static-file-serving/public/")

	output = string(conn.outBuf)

	if res.Headers["Content-Type"][0] != "text/html; charset=utf-8" {
		t.Errorf("Expected header Content-Type: text/html; charset=utf-8, but got %s", res.Headers["Content-Type"][0])
	}

	if !strings.Contains(output, "<h1>Hello from static home file!</h1>") {
		t.Errorf("Unexpected response body: %s", output)
	}

	res = &Res{
		Socket:  conn,
		Headers: make(map[string][]string),
	}

	res.Static("download.png", "./examples/static-file-serving/public/")

	output = string(conn.outBuf)

	if res.Headers["Content-Type"][0] != "image/png" {
		t.Errorf("Expected header image/png, but got %s", res.Headers["Content-Type"][0])
	}
}

func TestResponseCookies(t *testing.T) {

	res := &Res{
		Headers: map[string][]string{},
	}

	cookie := Cookie{
		Name:        "super",
		Value:       "cookie",
		Path:        "/",
		Domain:      "example.com",
		Expires:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		MaxAge:      86400,
		Secure:      true,
		HttpOnly:    true,
		SameSite:    "Lax",
		SessionOnly: true}

	res.SetCookie(cookie)

	cookies := res.Headers["Set-Cookie"]

	if len(cookies) != 1 {
		t.Errorf("Expected one cookies, but found %d", len(cookies))
	}

	cookieStr := cookies[0]
	parts := strings.Split(cookieStr, "; ")

	expectedParts := []string{
		"super=cookie",
		"Path=/",
		"Domain=example.com",
		"Expires=Wed, 01 Jan 2025 00:00:00 UTC",
		"Max-Age=86400",
		"Secure",
		"HttpOnly",
		"SameSite=Lax",
		"SessionOnly=true",
	}

	if len(parts) != len(expectedParts) {
		t.Errorf("Expected %d cookie parts, got %d", len(expectedParts), len(parts))
	}

	for i, part := range expectedParts {
		if parts[i] != part {
			t.Errorf("Part %d mismatch:\nExpected: %s\nGot:      %s", i, part, parts[i])
		}
	}

	expectedCookie := "super=cookie; Path=/; Domain=example.com; Expires=Wed, 01 Jan 2025 00:00:00 UTC; Max-Age=86400; Secure; HttpOnly; SameSite=Lax; SessionOnly=true"
	if cookieStr != expectedCookie {
		t.Errorf("Cookie string mismatch:\nExpected: %s\nGot:      %s", expectedCookie, cookieStr)
	}
}
