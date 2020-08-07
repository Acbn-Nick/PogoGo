package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	server "github.com/Acbn-Nick/pogogo/internal/server"
)

func main() {
	sigs := make(chan os.Signal, 1)

	s := server.New()

	go s.Start()

	<-sigs
	log.Info("Killing server")
}
