package main

import (
	"fmt"
	"log"

	"github.com/muhammadzkralla/zttp"
)

func main() {
	app := zttp.NewApp()

	app.Use(func(req *zttp.Req, res *zttp.Res, next func()) {
		res.Header("GlobalMiddleware", "true")
		log.Printf("m1: Request: %s %s\n", req.Method, req.Path)
		next()
	})

	app.Use("/home", func(req *zttp.Req, res *zttp.Res, next func()) {
		res.Header("HomeMiddleware", "true")
		log.Printf("m2: Request: %s %s\n\n", req.Method, req.Path)
		next()
	})

	router := app.NewRouter("/api/v1")

	router.Use(func(req *zttp.Req, res *zttp.Res, next func()) {
		res.Header("RouterMiddleware", "true")
		next()
	})

	router.Get("/home", func(req *zttp.Req, res *zttp.Res) {
		response := fmt.Sprintf("The request base url is: %s", req.Host())
		res.Status(200).Send(response)
	})

	router.Post("/home/:postId/comment/:commentId", func(req *zttp.Req, res *zttp.Res) {
		res.Status(201).Send("/api/v1/home post found with postId: " + req.Param("postId") + " and commentId: " + req.Param("commentId"))
	})

	app.Start(8080)
}
