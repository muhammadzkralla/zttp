package zttp

import (
	"strings"
	"testing"
)

// Test GET route matching
func TestGetRouteMatching(t *testing.T) {
	app := NewApp()

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
	app := NewApp()

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
	app := NewApp()

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
	app := NewApp()

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
	app := NewApp()

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

// Test dynamic routing
func TestDynamicRouting(t *testing.T) {
	app := NewApp()

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

// Test 404 not found handler
func TestNotFoundHandler(t *testing.T) {
	app := NewApp()

	// Perform a request to a non-existing handler
	response := mockRequest(app, "GET", "/test", "")

	if !strings.Contains(response, "Not Found") {
		t.Errorf("Expected 'Not Found', but got %s", response)
	}
}

func TestCustomRouter(t *testing.T) {
	app := NewApp()

	router := app.NewRouter("/api/v1")

	router.Get("/home", func(req Req, res Res) {
		res.Status(200).Send("/api/v1/home get found")
	})

	router.Post("/home/:postId/comment/:commentId", func(req Req, res Res) {
		res.Status(201).Send("/api/v1/home post found with postId: " + req.Param("postId") + " and commentId: " + req.Param("commentId"))
	})

	response := mockRequest(app, "GET", "/api/v1/home", "")

	if !strings.Contains(response, "/api/v1/home get found") {
		t.Errorf("Expected 'api/v1/home', but got %s", response)
	}

	response = mockRequest(app, "POST", "/api/v1/home/123/comment/comment1", "")

	if !strings.Contains(response, "/api/v1/home post found with postId: 123 and commentId: comment1") {
		t.Errorf("Expected 'api/v1/home', but got %s", response)
	}

}

// Test path cleaning logic
func TestCleanPath(t *testing.T) {
	tests := []struct {
		prefix string
		path   string
		want   string
	}{
		{"/api", "/users", "/api/users"},
		{"/api/", "/users", "/api/users"},
		{"/api", "users", "/api/users"},
		{"/api", "/users/", "/api/users"},
		{"/", "/users", "/users"},
		{"/", "users", "/users"},
		{"", "/users", "/users"},
		{"", "users", "/users"},
		{"/api", "//users//profile", "/api/users/profile"},
	}

	for _, tt := range tests {
		got := cleanPath(tt.prefix, tt.path)
		if got != tt.want {
			t.Errorf("cleanPath(%q, %q) = %q; want %q", tt.prefix, tt.path, got, tt.want)
		}
	}
}
