package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /point", postPointHandler)

	server := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}

	server.ListenAndServe()
	log.Printf("Server started on %s", server.Addr)
}
