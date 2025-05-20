package main

import (
	"log"
	"net/http"
	"time"

	"github.com/muhammadzkralla/zttp"
)

func main() {
	app := zttp.NewApp()

	app.Get("/fresh", func(req *zttp.Req, res *zttp.Res) {
		res.Header("ETag", "version1")

		if req.Fresh() {
			log.Println("Fresh")
			res.Status(304).End()
		} else {
			res.Status(200).Send("Not Fresh")
		}
	})

	app.Get("/last-modified", func(req *zttp.Req, res *zttp.Res) {
		res.Header("ETag", "version1")

		lastModified := time.Now().Truncate(24 * time.Hour).UTC()
		res.Header("Last-Modified", lastModified.Format(http.TimeFormat))

		if req.Fresh() {
			log.Println("Client's cached version is still valid")
			res.Status(304).End()
		} else {
			res.Status(200).Send("Fresh content generated at: " + time.Now().UTC().Format(time.RFC3339))
		}

	})

	app.Get("/no-cache-example", func(req *zttp.Req, res *zttp.Res) {
		res.Header("ETag", "static-version-123")

		if req.Fresh() {
			log.Println("Serving from cache")
			res.Status(304).End()
		} else {
			res.Status(200).Send("This always shows when no-cache is requested")
		}

	})

	app.Start(8080)
}
