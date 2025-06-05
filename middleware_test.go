package zttp

import (
	"strings"
	"testing"
)

// Test middleware
func TestMiddleware(t *testing.T) {
	app := NewApp()

	// Global middleware
	app.Use(func(req *Req, res *Res, next func()) {
		res.Send("GlobalMiddleware\n")
		next()
	})

	// Path-specific middleware
	app.Use("/api", func(req *Req, res *Res, next func()) {
		res.Send("ApiMiddleware\n")
		next()
	})

	// /test should only trigger global middleware
	app.Get("/test", func(req *Req, res *Res) {
		res.Send("Handler: /test")
	})

	// /api should trigger both global and /api middleware
	app.Get("/api", func(req *Req, res *Res) {
		res.Send("Handler: /api")
	})

	tests := []struct {
		name             string
		path             string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:             "global middleware only",
			path:             "/test",
			shouldContain:    []string{"GlobalMiddleware", "Handler: /test"},
			shouldNotContain: []string{"ApiMiddleware"},
		},
		{
			name:             "both global and path-specific middleware",
			path:             "/api",
			shouldContain:    []string{"GlobalMiddleware", "ApiMiddleware", "Handler: /api"},
			shouldNotContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := mockRequest(app, "GET", tt.path, "")

			for _, expected := range tt.shouldContain {
				if !strings.Contains(response, expected) {
					t.Errorf("Expected response to contain '%s' for path '%s', got: %s",
						expected, tt.path, response)
				}
			}

			for _, unexpected := range tt.shouldNotContain {
				if strings.Contains(response, unexpected) {
					t.Errorf("Expected response NOT to contain '%s' for path '%s', got: %s",
						unexpected, tt.path, response)
				}
			}
		})
	}
}

// Test app (global) middlewares matching vs router (local) middlewares matching
func TestRouterMiddlewares(t *testing.T) {
	app := NewApp()

	// Global middleware
	app.Use(func(req *Req, res *Res, next func()) {
		res.Header("GlobalMiddleware", "true")
		next()
	})

	router := app.NewRouter("/api/v1")

	// Router middleware
	router.Use(func(req *Req, res *Res, next func()) {
		res.Header("RouterMiddleware", "true")
		next()
	})

	router.Get("/test", func(req *Req, res *Res) {
		res.Status(200).Send("")
	})

	app.Get("/test", func(req *Req, res *Res) {
		res.Status(200).Send("")
	})

	tests := []struct {
		name             string
		path             string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:             "router with both middlewares",
			path:             "/api/v1/test",
			shouldContain:    []string{"GlobalMiddleware", "RouterMiddleware"},
			shouldNotContain: []string{},
		},
		{
			name:             "app route with only global middleware",
			path:             "/test",
			shouldContain:    []string{"GlobalMiddleware"},
			shouldNotContain: []string{"RouterMiddleware"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := mockRequest(app, "GET", tt.path, "")

			for _, expected := range tt.shouldContain {
				if !strings.Contains(response, expected) {
					t.Errorf("Expected response to contain '%s' for path '%s', got: %s",
						expected, tt.path, response)
				}
			}

			for _, unexpected := range tt.shouldNotContain {
				if strings.Contains(response, unexpected) {
					t.Errorf("Expected response NOT to contain '%s' for path '%s', got: %s",
						unexpected, tt.path, response)
				}
			}
		})
	}
}
