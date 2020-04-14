package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/x45e/televise"
)

func main() {
	flag.Parse()

	db := os.Getenv("TELEVISE_DB")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	cfg := televise.Config{
		Addr: addr,
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
