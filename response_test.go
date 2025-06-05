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

// Test sending response
func TestResponseMethods(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Res)
		contains []string
		headers  map[string]string
	}{
		{
			name: "Send plain text",
			setup: func(r *Res) {
				r.Send("OK")
			},
			contains: []string{
				"HTTP/1.1 200 OK",
				"Content-Type: text/plain",
				"OK",
			},
		},
		{
			name: "Send JSON from map",
			setup: func(r *Res) {
				r.Json(map[string]string{"message": "OK"})
			},
			contains: []string{
				"HTTP/1.1 200 OK",
				"Content-Type: application/json",
				`"message":"OK"`,
			},
		},
		{
			name: "Send JSON from struct",
			setup: func(r *Res) {
				r.Json(User{Name: "Zkrallah", Age: 21})
			},
			contains: []string{
				"HTTP/1.1 200 OK",
				"Content-Type: application/json",
				`"name":"Zkrallah"`,
				`"age":21`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &MockConn{}
			res := &Res{
				Socket:     conn,
				StatusCode: 200,
				Headers:    make(map[string][]string),
			}

			tt.setup(res)
			output := string(conn.outBuf)

			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("Expected response to contain '%s'", s)
				}
			}

			for k, v := range tt.headers {
				if res.Headers[k][0] != v {
					t.Errorf("Expected header %s: %s, got %s", k, v, res.Headers[k][0])
				}
			}
		})
	}
}

// Test setting response headers
func TestResponseHeaders(t *testing.T) {
	t.Run("Multiple headers", func(t *testing.T) {
		res := Res{Headers: make(map[string][]string)}
		res.Header("Header1", "header1")
		res.Header("Header1", "notheader1")
		res.Header("Header2", "header2")

		if len(res.Headers["Header1"]) != 2 || res.Headers["Header1"][0] != "header1" ||
			res.Headers["Header1"][1] != "notheader1" || res.Headers["Header2"][0] != "header2" {
			t.Error("Header setting failed")
		}
	})
}

// Test setting response status code
func TestResponseStatus(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"Status 500", 500, 500},
		{"Status 301", 301, 301},
		{"Status 404", 404, 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &Res{}
			res.Status(tt.code)
			if res.StatusCode != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, res.StatusCode)
			}
		})
	}
}

// Test static file serving
func TestStaticFileServing(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		contains string
		ctype    string
	}{
		{
			name:     "Serve HTML index",
			file:     "index.html",
			contains: "<h1>Hello from static index file!</h1>",
			ctype:    "text/html; charset=utf-8",
		},
		{
			name:     "Serve HTML home",
			file:     "home.html",
			contains: "<h1>Hello from static home file!</h1>",
			ctype:    "text/html; charset=utf-8",
		},
		{
			name: "Serve PNG image",
			file: "download.png",
			// Can't check binary content
			contains: "",
			ctype:    "image/png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &MockConn{}
			res := &Res{
				Socket:  conn,
				Headers: make(map[string][]string),
			}

			res.Static(tt.file, "./examples/static-file-serving/public/")
			output := string(conn.outBuf)

			if res.Headers["Content-Type"][0] != tt.ctype {
				t.Errorf("Expected Content-Type %s, got %s", tt.ctype, res.Headers["Content-Type"][0])
			}

			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("Expected response to contain '%s'", tt.contains)
			}
		})
	}
}

// Test setting response cookie
func TestResponseCookies(t *testing.T) {
	t.Run("Complex cookie", func(t *testing.T) {
		res := &Res{Headers: make(map[string][]string)}

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
			SessionOnly: true,
		}

		res.SetCookie(cookie)
		cookies := res.Headers["Set-Cookie"]

		if len(cookies) != 1 {
			t.Fatalf("Expected 1 cookie, got %d", len(cookies))
		}

		expected := []string{
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

		parts := strings.Split(cookies[0], "; ")
		if len(parts) != len(expected) {
			t.Fatalf("Expected %d cookie parts, got %d", len(expected), len(parts))
		}

		for i, part := range expected {
			if parts[i] != part {
				t.Errorf("Part %d mismatch:\nExpected: %s\nGot:      %s", i, part, parts[i])
			}
		}
	})
}
