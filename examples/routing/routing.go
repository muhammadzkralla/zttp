package main

import (
	"fmt"

	"github.com/muhammadzkralla/zttp"
)

func main() {
	app := zttp.NewApp()

	app.Get("/home", func(req *zttp.Req, res *zttp.Res) {
		res.Send("Hello, World!")
	})

	app.Get("/post/:postId/comment/:commentId", func(req *zttp.Req, res *zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		res.Send("Post ID: " + postId + ", Comment ID: " + commentId)
	})

	app.Delete("/home", func(req *zttp.Req, res *zttp.Res) {
		res.Send("Deleted home")
	})

	app.Delete("/post/:postId/comment/:commentId", func(req *zttp.Req, res *zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		res.Send("Deleted Post ID: " + postId + ", Comment ID: " + commentId)
	})

	app.Post("/home", func(req *zttp.Req, res *zttp.Res) {
		reqBody := req.Body
		response := "You sent: " + reqBody
		res.Status(201).Send(response)
	})

	app.Post("/post/:postId/comment/:commentId", func(req *zttp.Req, res *zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		response := fmt.Sprintf("Posted %s for post id %s and comment id %s", req.Body, postId, commentId)
		res.Status(201).Send(response)
	})

	app.Put("/home", func(req *zttp.Req, res *zttp.Res) {
		reqBody := req.Body
		response := "Updated home with: " + reqBody
		res.Status(201).Send(response)
	})

	app.Put("/post/:postId/comment/:commentId", func(req *zttp.Req, res *zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		response := fmt.Sprintf("Updated post id %s and comment id %s with %s", postId, commentId, req.Body)
		res.Status(201).Send(response)
	})

	app.Patch("/home", func(req *zttp.Req, res *zttp.Res) {
		reqBody := req.Body
		response := "Patched home with: " + reqBody
		res.Send(response)
	})

	app.Patch("/post/:postId/comment/:commentId", func(req *zttp.Req, res *zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		response := fmt.Sprintf("Patched post id %s and comment id %s with %s", postId, commentId, req.Body)
		res.Status(201).Send(response)
	})

	app.Start(8080)
}
