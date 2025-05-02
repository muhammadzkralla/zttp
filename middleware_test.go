package zttp

import (
	"strings"
	"testing"
)

// Test middleware
func TestMiddleware(t *testing.T) {
	app := &App{}

	// Global middleware
	app.Use(func(req Req, res Res, next func()) {
		res.Send("GlobalMiddleware\n")
		next()
	})

	// Path-specific middleware
	app.Use("/api", func(req Req, res Res, next func()) {
		res.Send("ApiMiddleware\n")
		next()
	})

	// /test should only trigger global middleware
	app.Get("/test", func(req Req, res Res) {
		res.Send("Handler: /test")
	})

	// /api should trigger both global and /api middleware
	app.Get("/api", func(req Req, res Res) {
		res.Send("Handler: /api")
	})

	// Test /test
	response1 := mockRequest(app, "GET", "/test", "")
	if !strings.Contains(response1, "GlobalMiddleware") {
		t.Errorf("Expected global middleware for /test, got: %s", response1)
	}
	if strings.Contains(response1, "ApiMiddleware") {
		t.Errorf("Did not expect /api middleware for /test, got: %s", response1)
	}

	// Test /api
	response2 := mockRequest(app, "GET", "/api", "")
	if !strings.Contains(response2, "GlobalMiddleware") || !strings.Contains(response2, "ApiMiddleware") {
		t.Errorf("Expected both middlewares for /api, got: %s", response2)
	}
}
