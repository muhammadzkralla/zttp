package main

import (
	"log"
	"time"

	"github.com/muhammadzkralla/zttp"
)

func main() {
	app := zttp.NewApp()

	app.Get("/set", func(req *zttp.Req, res *zttp.Res) {
		cookie := zttp.Cookie{
			Name:    "username",
			Value:   "zkrallah",
			Expires: time.Now().Add(15 * time.Second),
		}

		log.Printf("Cookie: %s", req.Cookies)

		res.SetCookie(cookie)
		res.Status(200).Send("Set")
	})

	app.Get("/clear", func(req *zttp.Req, res *zttp.Res) {
		log.Printf("Cookie: %s", req.Cookies)

		res.ClearCookie("username")
		res.Status(200).Send("Cleared")
	})

	app.Start(8080)
}
