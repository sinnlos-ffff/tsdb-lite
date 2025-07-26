package server

import (
	"encoding/json"
	"net/http"
	"time"
)

type PostTimeSeriesRequest struct {
	Metric string            `json:"metric"`
	Tags   map[string]string `json:"tags"`
}

func (s *Server) PostTimeSeriesHandler(w http.ResponseWriter, r *http.Request) {
	var data PostTimeSeriesRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.Db.AddTimeSeries(data.Metric, data.Tags)

	w.WriteHeader(http.StatusOK)
}

// TODO: Proper validation
type PostPointRequest struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags"`
}

// TODO: Accept multiple points in a single request
func (s *Server) PostPointHandler(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusOK)
}
