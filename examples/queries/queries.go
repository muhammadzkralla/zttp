package main

import "github.com/muhammadzkralla/zttp"

func main() {
	app := zttp.NewApp()

	app.Get("/user", func(req *zttp.Req, res *zttp.Res) {
		queries := req.Queries
		var response string

		for k, v := range queries {
			response = response + k + ": " + v + "\n"
		}
		res.Send(response)
	})

	app.Get("/query", func(req *zttp.Req, res *zttp.Res) {
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

	app.Start(8080)
}
