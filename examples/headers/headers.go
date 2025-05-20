package main

import "github.com/muhammadzkralla/zttp"

func main() {
	app := zttp.NewApp()

	app.Get("/get/header", func(req *zttp.Req, res *zttp.Res) {
		h1 := req.Header("Header1")
		h2 := req.Header("Header2")

		var response string

		response = response + "h1 value: " + h1 + "\n"
		response = response + "h2 value: " + h2 + "\n"

		res.Send(response)
	})

	app.Get("/set/header", func(req *zttp.Req, res *zttp.Res) {
		res.Header("Header1", "header1")
		res.Header("Header1", "notheader1")
		res.Header("Header2", "header2")

		res.Send("ok")
	})

	app.Start(8080)
}
