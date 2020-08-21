package main

import (
	"context"
	"runtime/debug"

	log "github.com/sirupsen/logrus"

	server "github.com/Acbn-Nick/pogogo/internal/client"
)

func main() {
	debug.SetGCPercent(10)
	ctx, cancel := context.WithCancel(context.Background())
	c, done := server.New(ctx)

	go c.Start()

	<-done
	log.Info("killing client")
	cancel()
}
