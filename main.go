package main

import (
	"log"
	"time"

	"github.com/sinnlos-ffff/tsdb-lite/server"
)

func main() {
	s := server.NewServer(&server.Config{
		// TODO: Find optimal compaction interval
		// TODO: Add test
		CompactionInterval: time.Minute,
	})
	port := s.HttpServer.Addr

	log.Printf("Starting server on %s\n", port)
	if err := s.HttpServer.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server on %s\n", port)
	}
}
