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

	app.Get("/home", func(req *zttp.Req, res *zttp.Res) {
		res.Status(200).Send("Get Home!")
	})

	app.Get("/nothome", func(req *zttp.Req, res *zttp.Res) {
		res.Status(200).Send("Get Not Home!")
	})

	app.Delete("/home", func(req *zttp.Req, res *zttp.Res) {
		res.Status(200).Send("Delete Home!")
	})

	app.Delete("/nothome", func(req *zttp.Req, res *zttp.Res) {
		res.Status(200).Send("Deleted Not Home!")
	})

	app.Post("/home", func(req *zttp.Req, res *zttp.Res) {
		reqBody := req.Body
		response := fmt.Sprintf("Post Home: %s", reqBody)
		res.Status(201).Send(response)
	})

	app.Post("/nothome", func(req *zttp.Req, res *zttp.Res) {
		reqBody := req.Body
		response := fmt.Sprintf("Post Not Home: %s", reqBody)
		res.Status(201).Send(response)
	})

	app.Put("/home", func(req *zttp.Req, res *zttp.Res) {
		reqBody := req.Body
		response := fmt.Sprintf("Put Home: %s", reqBody)
		res.Status(201).Send(response)
	})

	app.Put("/nothome", func(req *zttp.Req, res *zttp.Res) {
		reqBody := req.Body
		response := fmt.Sprintf("Put Not Home: %s", reqBody)
		res.Status(201).Send(response)
	})

	app.Patch("/home", func(req *zttp.Req, res *zttp.Res) {
		reqBody := req.Body
		response := fmt.Sprintf("Patch Home: %s", reqBody)
		res.Status(201).Send(response)
	})

	app.Patch("/nothome", func(req *zttp.Req, res *zttp.Res) {
		reqBody := req.Body
		response := fmt.Sprintf("Patch Not Home: %s", reqBody)
		res.Status(201).Send(response)
	})

	app.Start(8080)
}
