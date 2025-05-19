<h1 align="center"> ZTTP </h1>

## Introduction

**ZTTP** is a lightweight, zero-dependency backend framework written in Go, built directly over raw TCP sockets. Designed as a toy project for educational purposes, it draws inspiration from modern web frameworks like [Gofiber](https://gofiber.io) and [Express.js](https://expressjs.com).

This project follows the Front Controller design pattern, a widely adopted architectural approach in web frameworks such as Express.js, Spring Boot, and more.

All incoming TCP connections are funneled concurrently through a centralized request handling function `handleClient`, which performs the following responsibilities:

- Parses the HTTP request (method, path, params, queries, headers, body, cookies, etc.)
- Extracts query parameters and dynamic route parameters
- Matches the request to a registered route handler using method and path
- Delegates the request to the matched handler with a unified request/response context
- Manages errors, timeouts, and connection lifecycle centrally

By applying this pattern, the application enforces a clean separation of concerns, consistent request preprocessing, and centralized control over the request lifecycle. This design simplifies extensibility (e.g., adding middleware, authentication, logging) and improves maintainability as the application grows.

## Why ZTTP?

ZTTP was created as a deep-dive into how web frameworks work under the hood. From TCP socket handling and HTTP parsing to request routing and middleware architecture. It is a hands-on exercise in systems-level web development using Go, with minimal abstractions.

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
