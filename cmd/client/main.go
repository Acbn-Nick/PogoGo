package main

import (
	"context"

	log "github.com/sirupsen/logrus"

	server "github.com/Acbn-Nick/pogogo/internal/client"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c, done := server.New(ctx)

	go c.Start()

	<-done
	log.Info("killing client")
	cancel()
}
