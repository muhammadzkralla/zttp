package main

import "github.com/muhammadzkralla/zttp"

func main() {
	app := zttp.NewApp()

	app.Get("/tls", func(req *zttp.Req, res *zttp.Res) {
		res.Status(200).Send("Secured with HTTPS!")
	})

	app.StartTls(8080, "cert.pem", "key.pem")
}
