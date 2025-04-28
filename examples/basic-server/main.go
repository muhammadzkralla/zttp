package main

import (
	"github.com/muhammadzkralla/zttp"
	"log"
)

func main() {

	app := zttp.App{}

	app.Use(func(req zttp.Req, res zttp.Res, next func()) {
		log.Printf("m1: Request: %s %s\n", req.Method, req.Path)
		next()
	})

	app.Use(func(req zttp.Req, res zttp.Res, next func()) {
		log.Printf("m2: Request: %s %s\n\n", req.Method, req.Path)
		next()
	})

	app.Get("/home", func(req zttp.Req, res zttp.Res) {
		res.Send("Hello, World!")
	})

	app.Post("/home", func(req zttp.Req, res zttp.Res) {
		reqBody := req.Body
		response := "You sent: " + reqBody
		res.Send(response)
	})

	app.Get("/post/:postId/comment/:commentId", func(req zttp.Req, res zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		res.Send("Post ID: " + postId + ", Comment ID: " + commentId)
	})

	app.Start(1069)
}
