package main

import (
	"log"

	"github.com/sinnlos-ffff/tsdb-lite/server"
)

func main() {
	server := server.NewServer()
	server.HttpServer.ListenAndServe()
	log.Printf("Server started on %s", server.HttpServer.Addr)
}
