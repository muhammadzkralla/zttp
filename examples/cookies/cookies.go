package main

import (
	"log"
	"time"

	"github.com/muhammadzkralla/zttp"
)

func main() {
	app := zttp.NewApp()

	app.Get("/", func(req *zttp.Req, res *zttp.Res) {
		cookie := zttp.Cookie{
			Name:    "username",
			Value:   "zkrallah",
			Expires: time.Now().Add(5 * time.Second),
		}

		log.Printf("Cookie: %s", req.Cookies)

		res.SetCookie(cookie)
		res.Status(200).Send("Ok")
	})

	app.Start(8080)
}
