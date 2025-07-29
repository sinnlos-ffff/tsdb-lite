package main

import (
	"log"

	"github.com/sinnlos-ffff/tsdb-lite/server"
)

func main() {
	s := server.NewServer()
	port := s.HttpServer.Addr

	log.Printf("Starting server on %s\n", port)
	if err := s.HttpServer.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server on %s\n", port)
	}
}
