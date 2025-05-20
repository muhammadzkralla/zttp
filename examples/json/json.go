package main

import "github.com/muhammadzkralla/zttp"

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	app := zttp.NewApp()
	app.PrettyPrintJSON = true

	app.Post("/user", func(req *zttp.Req, res *zttp.Res) {
		var user User
		err := req.ParseJson(&user)
		if err != nil {
			res.StatusCode = 400
			res.Send("Invalid JSON")
			return
		}

		res.Status(201).Json(user)
	})

	app.Start(8080)
}
