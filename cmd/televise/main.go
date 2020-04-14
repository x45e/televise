package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/x45e/televise"
)

var (
	addr = flag.String("addr", ":8080", "HTTP listener address")
)

func main() {
	flag.Parse()

	db := os.Getenv("TELEVISE_DB")

	cfg := televise.Config{
		Addr: *addr,
		DB:   db,
	}

	app, err := televise.Start(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
