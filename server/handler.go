package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sinnlos-ffff/tsdb-lite/database"
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

	if err := s.Db.AddTimeSeries(data.Metric, data.Tags); err != nil {
		http.Error(w, "Failed to add time series: "+err.Error(), http.StatusBadRequest)
		return
	}

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

	if err := s.Db.AddPoint(data.Metric, data.Tags, data.Timestamp, data.Value); err != nil {
		http.Error(w, "Failed to add point: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type GetRangeRequest struct {
	Metric string            `json:"metric"`
	Tags   map[string]string `json:"tags"`
	Start  int64             `json:"start"`
	End    int64             `json:"end"`
}

type GetRangeResponse struct {
	Points []database.Point `json:"points"`
}

func (s *Server) GetRangeHandler(w http.ResponseWriter, r *http.Request) {
	var data GetRangeRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	points, err := s.Db.GetRange(data.Metric, data.Tags, data.Start, data.End)
	if err != nil {
		http.Error(w, "Failed to get range: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := GetRangeResponse{
		Points: points,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
