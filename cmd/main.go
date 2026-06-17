package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/reiyuchan/craftctl/internal/config"
	"github.com/reiyuchan/craftctl/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.New()
	app := server.New(cfg)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("Shutdown signal received")
		if err := app.Stop(); err != nil {
			log.Printf("Error stopping server: %v", err)
		}
		os.Exit(0)
	}()

	return app.Listen()
}
