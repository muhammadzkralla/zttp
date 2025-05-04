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
	app := zttp.NewApp()
	app.PrettyPrintJSON = true

	app.Use(func(req zttp.Req, res zttp.Res, next func()) {
		res.Header("GlobalMiddleware", "true")
		log.Printf("m1: Request: %s %s\n", req.Method, req.Path)
		next()
	})

	app.Use("/home", func(req zttp.Req, res zttp.Res, next func()) {
		res.Header("HomeMiddleware", "true")
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
		res.Status(201).Send(response)
	})

	app.Post("/post/:postId/comment/:commentId", func(req zttp.Req, res zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		response := fmt.Sprintf("Posted %s for post id %s and comment id %s", req.Body, postId, commentId)
		res.Status(201).Send(response)
	})

	app.Put("/home", func(req zttp.Req, res zttp.Res) {
		reqBody := req.Body
		response := "Updated home with: " + reqBody
		res.Status(201).Send(response)
	})

	app.Put("/post/:postId/comment/:commentId", func(req zttp.Req, res zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		response := fmt.Sprintf("Updated post id %s and comment id %s with %s", postId, commentId, req.Body)
		res.Status(201).Send(response)
	})

	app.Patch("/home", func(req zttp.Req, res zttp.Res) {
		reqBody := req.Body
		response := "Patched home with: " + reqBody
		res.Send(response)
	})

	app.Patch("/post/:postId/comment/:commentId", func(req zttp.Req, res zttp.Res) {
		postId := req.Params["postId"]
		commentId := req.Params["commentId"]
		response := fmt.Sprintf("Patched post id %s and comment id %s with %s", postId, commentId, req.Body)
		res.Status(201).Send(response)
	})

	app.Post("/user", func(req zttp.Req, res zttp.Res) {
		var user User
		err := req.ParseJson(&user)
		if err != nil {
			res.StatusCode = 400
			res.Send("Invalid JSON")
			return
		}

		res.Status(201).Json(user)
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

	app.Get("/get/header", func(req zttp.Req, res zttp.Res) {
		h1 := req.Header("Header1")
		h2 := req.Header("Header2")

		var response string

		response = response + "h1 value: " + h1 + "\n"
		response = response + "h2 value: " + h2 + "\n"

		res.Send(response)
	})

	app.Get("/set/header", func(req zttp.Req, res zttp.Res) {
		res.Header("Header1", "header1")
		res.Header("Header1", "notheader1")
		res.Header("Header2", "header2")

		res.Send("ok")
	})

	app.Get("/set/status", func(req zttp.Req, res zttp.Res) {
		res.Status(400).Send("Bad Request Manually")
	})

	router := app.NewRouter("/api/v1")

	router.Use(func(req zttp.Req, res zttp.Res, next func()) {
		res.Header("RouterMiddleware", "true")
		next()
	})

	router.Get("/home", func(req zttp.Req, res zttp.Res) {
		res.Status(200).Send("/api/v1/home get found")
	})

	router.Post("/home/:postId/comment/:commentId", func(req zttp.Req, res zttp.Res) {
		res.Status(201).Send("/api/v1/home post found with postId: " + req.Param("postId") + " and commentId: " + req.Param("commentId"))
	})

	app.Start(1069)
}
