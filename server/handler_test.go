package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sinnlos-ffff/tsdb-lite/database"
	"github.com/stretchr/testify/assert"
)

func TestPostTimeSeriesHandler(t *testing.T) {
	db := database.NewDatabase()
	server := &Server{Db: db}

	metric := "test_metric"
	tags := map[string]string{"tag1": "value1"}
	reqBody, _ := json.Marshal(PostTimeSeriesRequest{
		Metric: metric,
		Tags:   tags,
	})

	req, err := http.NewRequest("POST", "/timeseries", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(server.PostTimeSeriesHandler)
	handler.ServeHTTP(responseRecorder, req)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	key := database.GenerateKey(metric, tags)
	ts, ok := db.GetShard(key).Series[key]

	assert.True(t, ok, "TimeSeries not found for key: %s", key)
	assert.Equal(t, metric, ts.Metric)
	assert.Equal(t, tags, ts.Tags)

	// Re-posting a time series with the same metric and tags returns an error
	req2, err := http.NewRequest("POST", "/timeseries", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	responseRecorder2 := httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder2, req2)
	assert.Equal(t, http.StatusBadRequest, responseRecorder2.Code)
}

func TestPostPointHandler(t *testing.T) {
	db := database.NewDatabase()
	server := &Server{Db: db}

	metric := "test_metric"
	tags := map[string]string{"tag1": "value1"}
	timestamp := time.Now().Unix()
	value := 1.23

	// Add a time series to the database
	err := db.AddTimeSeries(metric, tags)
	assert.NoError(t, err)

	reqBody, _ := json.Marshal(PostPointRequest{
		Metric:    metric,
		Timestamp: timestamp,
		Value:     value,
		Tags:      tags,
	})

	req, err := http.NewRequest("POST", "/point", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(server.PostPointHandler)
	handler.ServeHTTP(responseRecorder, req)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	// Check if the point was added to the time series
	key := database.GenerateKey(metric, tags)
	ts, ok := db.GetShard(key).Series[key]
	assert.True(t, ok)
	assert.Equal(t, 1, len(ts.Chunks[0].Points))
	assert.Equal(t, timestamp, ts.Chunks[0].Points[0].Timestamp)
	assert.Equal(t, value, ts.Chunks[0].Points[0].Value)

	// Posting a point with a future timestamp returns an error
	futureTimestamp := time.Now().Unix() + 1000
	reqBody, _ = json.Marshal(PostPointRequest{
		Metric:    metric,
		Timestamp: futureTimestamp,
		Value:     1.23,
		Tags:      tags,
	})

	req, err = http.NewRequest("POST", "/point", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, req)

	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)
}

func TestGetRangeHandler(t *testing.T) {
	db := database.NewDatabase()
	server := &Server{Db: db}

	metric := "test_metric"
	tags := map[string]string{"tag1": "value1"}
	err := db.AddTimeSeries(metric, tags)
	assert.NoError(t, err)

	// Add points to the time series
	timestamp1 := int64(1000)
	value1 := 10.5
	db.AddPoint(metric, tags, timestamp1, value1)

	timestamp2 := int64(2000)
	value2 := 20.5
	db.AddPoint(metric, tags, timestamp2, value2)

	timestamp3 := int64(3000)
	value3 := 30.5
	db.AddPoint(metric, tags, timestamp3, value3)

	reqBody, _ := json.Marshal(GetRangeRequest{
		Metric: metric,
		Tags:   tags,
		Start:  0,
		End:    4000,
	})

	req, err := http.NewRequest("GET", "/range", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(server.GetRangeHandler)
	handler.ServeHTTP(responseRecorder, req)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var response GetRangeResponse
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Points, 3)

	// Test range that includes only the first two points
	reqBody, _ = json.Marshal(GetRangeRequest{
		Metric: metric,
		Tags:   tags,
		Start:  500,
		End:    2500,
	})

	req, err = http.NewRequest("GET", "/range", bytes.NewBuffer(reqBody))
	assert.NoError(t, err)

	responseRecorder = httptest.NewRecorder()
	handler.ServeHTTP(responseRecorder, req)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	err = json.Unmarshal(responseRecorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Points, 2)
	assert.Equal(t, value1, response.Points[0].Value)
	assert.Equal(t, value2, response.Points[1].Value)
}
