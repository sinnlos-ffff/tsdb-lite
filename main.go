package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// TODO: Proper validation
type PostPointRequest struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags"`
}

// TODO: Accept multiple points in a single request
func postPointHandler(w http.ResponseWriter, r *http.Request) {
	var data PostPointRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Remove if it turns out to be unnecessary
	if data.Timestamp > time.Now().Unix() {
		http.Error(w, "Timestamp cannot be in the future", http.StatusBadRequest)
		return
	}

	// TODO: Save metric data
	log.Printf("got metric: %+v", data)

	w.WriteHeader(http.StatusOK)
}

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
