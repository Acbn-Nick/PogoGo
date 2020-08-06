package main

import (
	api "github.com/Acbn-Nick/pogogo/api"
	server "github.com/Acbn-Nick/pogogo/internal/server"
)

func main() {
	var ()

	s := server.NewServer()

	s.Upload(&api.UploadRequest{Password: "asd"})
}
