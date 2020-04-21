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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	cfg := televise.Config{
		Addr: addr,
		DB: televise.DBConfig{
			Addr:     os.Getenv("TELEVISE_CASSANDRA_ADDR"),
			Username: os.Getenv("TELEVISE_CASSANDRA_USER"),
			Password: os.Getenv("TELEVISE_CASSANDRA_PASSWORD"),
		},
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
