package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	server "github.com/Acbn-Nick/pogogo/internal/server"
)

func main() {
	cleanup := flag.Bool("c", false, "runs cleanup at program start")

	flag.Parse()

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	s, done := server.New(ctx, *cleanup)

	go s.Start()

	<-sigs
	log.Info("killing server")
	cancel()
	<-done
}
