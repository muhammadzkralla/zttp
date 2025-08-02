<h1 align="center"> ZTTP </h1>

## Introduction

**ZTTP** is a minimal, zero-dependency, extremely fast backend framework written in Go, built directly over raw TCP sockets. Designed as a toy project for educational purposes, it draws inspiration from modern web frameworks like [Gofiber](https://gofiber.io) and [Express.js](https://expressjs.com).

This project follows the Front Controller design pattern, a widely adopted architectural approach in web frameworks such as [Spring Boot](https://spring.io), [Express.js](https://expressjs.com), and more.

All incoming TCP connections are funneled concurrently through a centralized request handling function `handleClient()`, which performs the following responsibilities:

- Parses the HTTP request (method, path, params, queries, headers, body, cookies, etc.)
- Extracts query parameters and dynamic route parameters
- Matches the request to a registered route handler using method and path
- Delegates the request to the matched handler with a unified request/response context
- Centrally Manages errors, timeouts, middlewares, connection lifecycle, and more.

By applying this pattern, the application enforces a clean separation of concerns, consistent request preprocessing, and centralized control over the request lifecycle. This design simplifies extensibility (e.g., adding middleware, authentication, logging) and improves maintainability as the application scales.

To state some numbers, I tested the same routes and benchmarks with different frameworks using wrk and took the average:

- 300k RPS, 3.5 ms latency using GoFiber
- 135k RPS, 8.7 ms latency using ZTTP
- 67k RPS, 34 ms latency using Spring WebMVC
- 55k RPS, 19 ms latency using Spring WebFlux
- 10k RPS, 135 ms latency using Express.js (Node)
- 1.7k RPS, 128 ms latency using Flask

Benchmarks included different core numbers, time periods, routes, etc, **all on the same machine separately**, and those are the average values.

## Why ZTTP?

ZTTP was created as a deep-dive into how web frameworks work under the hood. From TCP socket handling and HTTP parsing to request routing and middleware architecture. It is a hands-on exercise in systems-level web development using Go, with minimal abstractions.

I decided not to use any external HTTP engines, not even Go's `net/http` standard library, and handle all the logic from scratch starting from the TCP layer.

Everything in this project is perfectly aligned with the RFC standards and HTTP/1.1 structure, as I spent days reading the RFC standards specific to each feature before starting to implement it.

Whether you're learning how the web works or exploring Go's networking capabilities, ZTTP is designed to be small enough to understand yet expressive enough to grow.

## Installation

To use ZTTP in your Go project, simply run:
```bash
go get github.com/muhammadzkralla/zttp
```

Then, import it in your code:
```go
import "github.com/muhammadzkralla/zttp"
```

> [!NOTE]
> ZTTP is still under active development and may lack some features. It fetches the latest commit from the master branch. Semantic versioning and releases will be introduced later when I feel it's ready.

## Usage

Here’s a minimal example of how to spin up a simple ZTTP server:

```go
package main

import (
	"github.com/muhammadzkralla/zttp"
)

func main() {
	app := zttp.NewApp()

	app.Get("/", func(req *zttp.Req, res *zttp.Res) {
		res.Status(200).Send("Hello from ZTTP!")
	})

	app.Start(8080)
}
```

You can now test your server like this:

```bash
curl "localhost:8080"
```

You will get this printed in your terminal:

```bash
Hello from ZTTP!
```

## Features

### Core Functionality

- Raw TCP HTTP/1.1 server with concurrent connection handling
- Front Controller Design Pattern implementation
- Zero dependency (pure GO standard library)

### Routing

```go
app.Get("/path", handler)
app.Post("/path", handler)
app.Put("/path", handler)
app.Patch("/path", handler)
app.Delete("/path", handler)
```

### Path Parameters

```go
// Route: "/post/:postId/comment/:commentId"
params := req.Params                // All params (map[string]string)
postId := req.Param("postId")       // postId param
commentId := req.Param("commentId")  // commentId param
```

### Queries Parameters

```go
// URL: /user?name=John&age=30
queries := req.Queries              // All queries (map[string]string)
name := req.Query("name")           // name query
age := req.Query("age")           // age query
```

### Request Handling

- Body parsing: `req.Body` (raw string)

```go
body := req.Body    // raw string request body
```

- JSON parsing:

```go
// Parse request body into JSON and store the result in user variable
var user User
err := req.ParseJson(&user)
```
- Form data & file uploads:

```go
value := req.FormValue("field")     // Get `field` part from request
file, err := req.FormFile("file")   // Get `file` part from request
err = req.Save(file, "./uploads")   // Save file to disk in `./uploads` directory
```

- Accept headers processing:

```go
log.Println(req.Accepts("html"))                         // "html"
log.Println(req.AcceptsCharsets("utf-16", "iso-8859-1")) // "iso-8859-1"
log.Println(req.AcceptsEncodings("compress", "br"))      // "compress"
log.Println(req.AcceptsLanguages("en", "nl", "ru"))      // "nl"
```

And more request processing utilities.

### Response Handling

```go
res.Status(201).Send("text")        // Text response
res.Status(200).Json(data)          // JSON response
res.Status(304).End()               // Empty response
```

### Headers

```go
req.Headers                         // All request headers (map[string]string)
res.Headers                         // All response headers (map[string][]string)
req.Header("Header-Name")           // Get request header
res.Header("Key", "Value")         // Set response header
```

### Cookies

```go
req.Cookies                     // All request cookies (map[string]string)
res.SetCookie(zttp.Cookie{      // Set response cookie
    Name: "session",
    Value: "token",
    Expires: time.Now().Add(24*time.Hour)
})
```

### Static File Serving

```go
res.Static("index.html", "./public")         // Serve HTML file
res.Static("image.png", "./assets")         // Serve image file
```

### Middleware

```go
// Global middleware
app.Use(func(req *zttp.Req, res *zttp.Res, next func()) {
    // Pre-processing
    next()
    // Post-processing
})

// Route-specific middleware
app.Use("/path", middlewareHandler)
```

### Sub-Routers

```go
router := app.NewRouter("/api/v1")
router.Get("/endpoint", handler)    // Handles /api/v1/endpoint
router.Use("/path", middlewareHandler) // Router-specific middleware
```

### Cache Control

```go
res.Header("ETag", "version1")
res.Header("Last-Modified", timestamp)

// If the request is still fresh in the client's cache
if req.Fresh() {
    // Handle cached responses, return 304 Not Changed response
    res.Status(304).End()
}
```

### Error Handling

- Automatic panic recovery
- Manual error responses:

```go
res.Status(400).Send("Bad request")
```

### HTTPS Support via TLS

```go
// Start secure server with TLS
app.StartTLS(443, "cert.pem", "key.pem")
```

For More details, visit the [examples](https://github.com/muhammadzkralla/zttp/tree/master/examples/) section.
