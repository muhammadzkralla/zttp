package main

import (
	"log"

	"github.com/muhammadzkralla/zttp"
)

func main() {
	app := zttp.NewApp()

	app.Get("/accepts", func(req *zttp.Req, res *zttp.Res) {
		log.Println(req.Accepts("html"))                           // "html"
		log.Println(req.Accepts("text/html"))                      // "text/html"
		log.Println(req.Accepts("json", "text"))                   // "json"
		log.Println(req.Accepts("application/json"))               // "application/json"
		log.Println(req.Accepts("text/plain", "application/json")) // "application/json", due to quality
		log.Println(req.Accepts("image/png"))                      // ""
		log.Println(req.Accepts("png"))                            // ""

		res.Status(200).Send("Ok")
	})

	// Accept-Charset: utf-8, iso-8859-1;q=0.2
	// Accept-Encoding: gzip, compress;q=0.2
	// Accept-Language: en;q=0.8, nl, ru
	app.Get("/", func(req *zttp.Req, res *zttp.Res) {
		log.Println(req.AcceptsCharsets("utf-16", "iso-8859-1")) // "iso-8859-1"
		log.Println(req.AcceptsEncodings("compress", "br"))      // "compress"
		log.Println(req.AcceptsLanguages("en", "nl", "ru"))      // "nl"

		res.Status(200).Send("Ok")
	})

	app.Start(8080)
}
