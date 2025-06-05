package zttp

import (
	"strings"
	"testing"
)

// Test GET route matching
func TestRouteMatching(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		handler  func(*Req, *Res)
		expected string
	}{
		{
			name:   "GET route",
			method: "GET",
			path:   "/test",
			handler: func(req *Req, res *Res) {
				res.Send("GET route matched")
			},
			expected: "GET route matched",
		},
		{
			name:   "DELETE route",
			method: "DELETE",
			path:   "/test",
			handler: func(req *Req, res *Res) {
				res.Send("DELETE route matched")
			},
			expected: "DELETE route matched",
		},
		{
			name:   "POST route",
			method: "POST",
			path:   "/test",
			handler: func(req *Req, res *Res) {
				res.Send("POST route matched")
			},
			expected: "POST route matched",
		},
		{
			name:   "PUT route",
			method: "PUT",
			path:   "/test",
			handler: func(req *Req, res *Res) {
				res.Send("PUT route matched")
			},
			expected: "PUT route matched",
		},
		{
			name:   "PATCH route",
			method: "PATCH",
			path:   "/test",
			handler: func(req *Req, res *Res) {
				res.Send("PATCH route matched")
			},
			expected: "PATCH route matched",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()

			switch tt.method {
			case "GET":
				app.Get(tt.path, tt.handler)
			case "DELETE":
				app.Delete(tt.path, tt.handler)
			case "POST":
				app.Post(tt.path, tt.handler)
			case "PUT":
				app.Put(tt.path, tt.handler)
			case "PATCH":
				app.Patch(tt.path, tt.handler)
			}

			response := mockRequest(app, tt.method, tt.path, "")
			if !strings.Contains(response, tt.expected) {
				t.Errorf("Expected response to contain '%s', but got '%s'", tt.expected, response)
			}
		})
	}
}

// Test dynamic routing
func TestDynamicRouting(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		setup    func(*App)
		expected string
	}{
		{
			name:   "GET with params",
			method: "GET",
			path:   "/test/123/comment/comment1",
			setup: func(a *App) {
				a.Get("/test/:postId/comment/:commentId", func(req *Req, res *Res) {
					postId := req.Params["postId"]
					commentId := req.Params["commentId"]
					res.Send("GET: Post ID: " + postId + ", Comment ID: " + commentId)
				})
			},
			expected: "GET: Post ID: 123, Comment ID: comment1",
		},
		{
			name:   "POST with params",
			method: "POST",
			path:   "/test/456/comment/comment2",
			setup: func(a *App) {
				a.Post("/test/:postId/comment/:commentId", func(req *Req, res *Res) {
					postId := req.Params["postId"]
					commentId := req.Params["commentId"]
					res.Send("POST: Post ID: " + postId + ", Comment ID: " + commentId)
				})
			},
			expected: "POST: Post ID: 456, Comment ID: comment2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()
			tt.setup(app)

			response := mockRequest(app, tt.method, tt.path, "")

			if !strings.Contains(response, tt.expected) {
				t.Errorf("Expected response to contain '%s', but got '%s'", tt.expected, response)
			}
		})
	}
}

// Test 404 not found handler
func TestNotFoundHandler(t *testing.T) {
	t.Run("Non-existent route", func(t *testing.T) {
		app := NewApp()
		response := mockRequest(app, "GET", "/nonexistent", "")
		if !strings.Contains(response, "Not Found") {
			t.Errorf("Expected 'Not Found', but got '%s'", response)
		}
	})
}

// Test creating a custom router
func TestCustomRouter(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		setup    func(router *Router)
		expected string
	}{
		{
			name:   "GET from router",
			method: "GET",
			path:   "/api/v1/home",
			setup: func(router *Router) {
				router.Get("/home", func(req *Req, res *Res) {
					res.Status(200).Send("/api/v1/home get found")
				})
			},
			expected: "/api/v1/home get found",
		},
		{
			name:   "POST with params from router",
			method: "POST",
			path:   "/api/v1/home/123/comment/comment1",
			setup: func(router *Router) {
				router.Post("/home/:postId/comment/:commentId", func(req *Req, res *Res) {
					res.Status(201).Send("/api/v1/home post found with postId: " + req.Param("postId") + " and commentId: " + req.Param("commentId"))
				})
			},
			expected: "/api/v1/home post found with postId: 123 and commentId: comment1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp()

			router := app.NewRouter("/api/v1")

			tt.setup(router)
			response := mockRequest(app, tt.method, tt.path, "")
			if !strings.Contains(response, tt.expected) {
				t.Errorf("Expected '%s', but got '%s'", tt.expected, response)
			}
		})
	}
}

// Test path cleaning logic
func TestCleanPath(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		path   string
		want   string
	}{
		{
			name:   "basic join",
			prefix: "/api",
			path:   "/users",
			want:   "/api/users",
		},
		{
			name:   "trailing slash in prefix",
			prefix: "/api/",
			path:   "/users",
			want:   "/api/users",
		},
		{
			name:   "no leading slash in path",
			prefix: "/api",
			path:   "users",
			want:   "/api/users",
		},
		{
			name:   "trailing slash in path",
			prefix: "/api",
			path:   "/users/",
			want:   "/api/users",
		},
		{
			name:   "root prefix",
			prefix: "/",
			path:   "/users",
			want:   "/users",
		},
		{
			name:   "empty prefix",
			prefix: "",
			path:   "/users",
			want:   "/users",
		},
		{
			name:   "multiple slashes",
			prefix: "/api",
			path:   "//users//profile",
			want:   "/api/users/profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanPath(tt.prefix, tt.path)
			if got != tt.want {
				t.Errorf("cleanPath(%q, %q) = %q; want %q", tt.prefix, tt.path, got, tt.want)
			}
		})
	}
}
