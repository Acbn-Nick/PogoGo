package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	server "github.com/Acbn-Nick/pogogo/internal/client"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	c, done := server.New(ctx)

	go c.Start()

	//<-sigs
	<-done
	log.Info("killing client")
	cancel()
}
