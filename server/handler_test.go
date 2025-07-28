package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
	ts, ok := db.Series[key]

	assert.True(t, ok, "TimeSeries not found for key: %s", key)
	assert.Equal(t, metric, ts.Metric)
	assert.Equal(t, tags, ts.Tags)
}
