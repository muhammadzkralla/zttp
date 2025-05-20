package main

import "github.com/muhammadzkralla/zttp"

func main() {
	app := zttp.NewApp()

	app.Get("/panic", func(req *zttp.Req, res *zttp.Res) {
		panic("something went very wrong")
	})

	app.Start(8080)
}
