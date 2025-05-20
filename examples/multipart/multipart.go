package main

import (
	"fmt"
	"log"

	"github.com/muhammadzkralla/zttp"
)

func main() {
	app := zttp.NewApp()

	app.Post("/multipart", func(req *zttp.Req, res *zttp.Res) {
		part1 := req.FormValue("part1")
		part2, err := req.FormFile("file1")
		if err != nil {
			log.Println("err:", err)
			res.Status(400).Send("File upload error")
			return
		}

		part3 := req.FormValue("part2")
		part4, err := req.FormFile("file2")
		if err != nil {
			log.Println("err:", err)
			res.Status(400).Send("File upload error")
			return
		}

		err = req.Save(part2, "./uploads")
		if err != nil {
			log.Println("Failed to save file:", err)
			res.Status(500).Send("File saving error")
			return
		}

		err = req.Save(part4, "./uploads")
		if err != nil {
			log.Println("Failed to save file:", err)
			res.Status(500).Send("File saving error")
			return
		}

		response := fmt.Sprintf("Received part1: %s, Received part2: %s, Saved file1: %s, Saved file2: %s", part1, part3, part2.Filename, part4.Filename)

		res.Send(response)
	})

	app.Start(8080)
}
