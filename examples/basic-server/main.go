package main

import (
	"fmt"
	"log"

	"github.com/muhammadzkralla/zttp"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	app := zttp.App{
		PrettyPrintJSON: true,
	}

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

	app.Get("/post/:postId/comment/:commentId", func(req zttp.Req, res zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		res.Send("Post ID: " + postId + ", Comment ID: " + commentId)
	})

	app.Delete("/home", func(req zttp.Req, res zttp.Res) {
		res.Send("Deleted home")
	})

	app.Delete("/post/:postId/comment/:commentId", func(req zttp.Req, res zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		res.Send("Deleted Post ID: " + postId + ", Comment ID: " + commentId)
	})

	app.Post("/home", func(req zttp.Req, res zttp.Res) {
		reqBody := req.Body
		response := "You sent: " + reqBody
		res.Status = 201
		res.Send(response)
	})

	app.Post("/post/:postId/comment/:commentId", func(req zttp.Req, res zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		response := fmt.Sprintf("Posted %s for post id %s and comment id %s", req.Body, postId, commentId)
		res.Status = 201
		res.Send(response)
	})

	app.Put("/home", func(req zttp.Req, res zttp.Res) {
		reqBody := req.Body
		response := "Updated home with: " + reqBody
		res.Status = 201
		res.Send(response)
	})

	app.Put("/post/:postId/comment/:commentId", func(req zttp.Req, res zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		response := fmt.Sprintf("Updated post id %s and comment id %s with %s", postId, commentId, req.Body)
		res.Status = 201
		res.Send(response)
	})

	app.Patch("/home", func(req zttp.Req, res zttp.Res) {
		reqBody := req.Body
		response := "Patched home with: " + reqBody
		res.Status = 201
		res.Send(response)
	})

	app.Patch("/post/:postId/comment/:commentId", func(req zttp.Req, res zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		response := fmt.Sprintf("Patched post id %s and comment id %s with %s", postId, commentId, req.Body)
		res.Status = 201
		res.Send(response)
	})

	app.Post("/user", func(req zttp.Req, res zttp.Res) {
		var user User
		err := req.ParseJson(&user)
		if err != nil {
			res.Status = 400
			res.Send("Invalid JSON")
			return
		}

		res.Json(user)
	})

	app.Get("/user", func(req zttp.Req, res zttp.Res) {
		queries := req.Queries
		var response string

		for k, v := range queries {
			response = response + k + ": " + v + "\n"
		}
		res.Send(response)
	})

	app.Get("/query", func(req zttp.Req, res zttp.Res) {
		q1 := req.Query("userId")
		q2 := req.Query("name")
		q3 := req.Query("category")
		q4 := req.Query("unkown")

		var response string

		response = response + "q1 value: " + q1 + "\n"
		response = response + "q2 value: " + q2 + "\n"
		response = response + "q3 value: " + q3 + "\n"
		response = response + "q4 value: " + q4 + "\n"

		res.Send(response)
	})

	app.Start(1069)
}
