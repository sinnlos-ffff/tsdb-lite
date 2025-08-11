package main

import (
	"log"
	"time"

	"github.com/sinnlos-ffff/tsdb-lite/metrics"
	"github.com/sinnlos-ffff/tsdb-lite/server"
)

func main() {
	s := server.NewServer(&server.Config{
		CompactionInterval: time.Minute,
	})

	metrics.InitMetrics()

	port := ":8080"
	log.Printf("Starting gRPC server on %s\n", port)
	if err := s.ListenAndServe(port); err != nil {
		log.Fatalf("Failed to start server on %s: %v\n", port, err)
	}
}
