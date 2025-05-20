package main

import "github.com/muhammadzkralla/zttp"

func main() {
	app := zttp.NewApp()

	app.Get("/static/index.html", func(req *zttp.Req, res *zttp.Res) {
		res.Static("", "./public")
	})

	app.Get("/static/home.html", func(req *zttp.Req, res *zttp.Res) {
		res.Static("home.html", "./public")
	})

	app.Get("/static/download.png", func(req *zttp.Req, res *zttp.Res) {
		res.Static("download.png", "./public")
	})

	app.Start(8080)
}
