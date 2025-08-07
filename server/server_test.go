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
	"github.com/stretchr/testify/require"
)

func TestServer_Integration(t *testing.T) {
	// Create a new server
	server := NewServer()

	// Create a test server that uses the actual HTTP mux
	testServer := httptest.NewServer(server.HttpServer.Handler)
	defer testServer.Close()

	t.Run("POST /timeseries", func(t *testing.T) {
		metric := "cpu_usage"
		tags := map[string]string{"host": "server1", "region": "us-west"}

		reqBody := PostTimeSeriesRequest{
			Metric: metric,
			Tags:   tags,
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		// Make actual HTTP request
		resp, err := http.Post(
			testServer.URL+"/timeseries",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify the time series was actually created in the database
		key := database.GenerateKey(metric, tags)
		ts, ok := server.Db.GetShard(key).Series[key]
		assert.True(t, ok)
		assert.Equal(t, metric, ts.Metric)
		assert.Equal(t, tags, ts.Tags)
	})

	t.Run("POST /point", func(t *testing.T) {
		metric := "disk_usage"
		tags := map[string]string{"host": "server3", "mount": "/var"}
		timestamp := time.Now().Unix()
		value := 75.5

		// First create the time series
		tsReqBody := PostTimeSeriesRequest{
			Metric: metric,
			Tags:   tags,
		}
		tsBody, err := json.Marshal(tsReqBody)
		require.NoError(t, err)

		tsResp, err := http.Post(
			testServer.URL+"/timeseries",
			"application/json",
			bytes.NewBuffer(tsBody),
		)
		require.NoError(t, err)
		defer tsResp.Body.Close()
		require.Equal(t, http.StatusOK, tsResp.StatusCode)

		// Now add a point
		pointReqBody := PostPointRequest{
			Metric:    metric,
			Timestamp: timestamp,
			Value:     value,
			Tags:      tags,
		}

		pointBody, err := json.Marshal(pointReqBody)
		require.NoError(t, err)

		resp, err := http.Post(
			testServer.URL+"/point",
			"application/json",
			bytes.NewBuffer(pointBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify the point was added
		key := database.GenerateKey(metric, tags)
		ts, ok := server.Db.GetShard(key).Series[key]
		require.True(t, ok)
		require.Len(t, ts.Chunks, 1)
		require.Len(t, ts.Chunks[0].Points, 1)
		assert.Equal(t, timestamp, ts.Chunks[0].Points[0].Timestamp)
		assert.Equal(t, value, ts.Chunks[0].Points[0].Value)
	})

	t.Run("GET /range", func(t *testing.T) {
		metric := "temperature"
		tags := map[string]string{"sensor": "A1", "location": "datacenter"}

		// Create time series
		tsReqBody := PostTimeSeriesRequest{
			Metric: metric,
			Tags:   tags,
		}
		tsBody, err := json.Marshal(tsReqBody)
		require.NoError(t, err)

		tsResp, err := http.Post(
			testServer.URL+"/timeseries",
			"application/json",
			bytes.NewBuffer(tsBody),
		)
		require.NoError(t, err)
		defer tsResp.Body.Close()
		require.Equal(t, http.StatusOK, tsResp.StatusCode)

		// Add multiple points
		points := []struct {
			timestamp int64
			value     float64
		}{
			{1000, 20.5},
			{2000, 21.0},
			{3000, 19.8},
			{4000, 22.3},
		}

		for _, point := range points {
			pointReqBody := PostPointRequest{
				Metric:    metric,
				Timestamp: point.timestamp,
				Value:     point.value,
				Tags:      tags,
			}

			pointBody, err := json.Marshal(pointReqBody)
			require.NoError(t, err)

			resp, err := http.Post(
				testServer.URL+"/point",
				"application/json",
				bytes.NewBuffer(pointBody),
			)
			require.NoError(t, err)
			resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)
		}

		// Query range
		rangeReqBody := GetRangeRequest{
			Metric: metric,
			Tags:   tags,
			Start:  0,
			End:    5000,
		}

		rangeBody, err := json.Marshal(rangeReqBody)
		require.NoError(t, err)

		// Note: Using a custom client to send GET with body since http.Get doesn't support it
		client := &http.Client{}
		req, err := http.NewRequest("GET", testServer.URL+"/range", bytes.NewBuffer(rangeBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var rangeResp GetRangeResponse
		err = json.NewDecoder(resp.Body).Decode(&rangeResp)
		require.NoError(t, err)

		assert.Len(t, rangeResp.Points, 4)

		// Verify points are returned in order
		for i, expectedPoint := range points {
			assert.Equal(t, expectedPoint.timestamp, rangeResp.Points[i].Timestamp)
			assert.Equal(t, expectedPoint.value, rangeResp.Points[i].Value)
		}
	})

	t.Run("GET /range - partial range", func(t *testing.T) {
		metric := "pressure"
		tags := map[string]string{"gauge": "B2"}

		// Create time series and add points (similar to above)
		tsReqBody := PostTimeSeriesRequest{
			Metric: metric,
			Tags:   tags,
		}
		tsBody, err := json.Marshal(tsReqBody)
		require.NoError(t, err)

		tsResp, err := http.Post(
			testServer.URL+"/timeseries",
			"application/json",
			bytes.NewBuffer(tsBody),
		)
		require.NoError(t, err)
		defer tsResp.Body.Close()
		require.Equal(t, http.StatusOK, tsResp.StatusCode)

		// Add points
		timestamps := []int64{1000, 2000, 3000, 4000}
		values := []float64{100.0, 105.0, 95.0, 110.0}

		for i, ts := range timestamps {
			pointReqBody := PostPointRequest{
				Metric:    metric,
				Timestamp: ts,
				Value:     values[i],
				Tags:      tags,
			}

			pointBody, err := json.Marshal(pointReqBody)
			require.NoError(t, err)

			resp, err := http.Post(
				testServer.URL+"/point",
				"application/json",
				bytes.NewBuffer(pointBody),
			)
			require.NoError(t, err)
			resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)
		}

		// Query partial range (should return only middle two points)
		rangeReqBody := GetRangeRequest{
			Metric: metric,
			Tags:   tags,
			Start:  1500, // Between first and second point
			End:    3500, // Between third and fourth point
		}

		rangeBody, err := json.Marshal(rangeReqBody)
		require.NoError(t, err)

		client := &http.Client{}
		req, err := http.NewRequest("GET", testServer.URL+"/range", bytes.NewBuffer(rangeBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var rangeResp GetRangeResponse
		err = json.NewDecoder(resp.Body).Decode(&rangeResp)
		require.NoError(t, err)

		// Should return only the middle two points
		assert.Len(t, rangeResp.Points, 2)
		assert.Equal(t, int64(2000), rangeResp.Points[0].Timestamp)
		assert.Equal(t, 105.0, rangeResp.Points[0].Value)
		assert.Equal(t, int64(3000), rangeResp.Points[1].Timestamp)
		assert.Equal(t, 95.0, rangeResp.Points[1].Value)
	})

	t.Run("Invalid routes return 404", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/invalid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Wrong HTTP method returns 405", func(t *testing.T) {
		// GET to /timeseries should fail (only POST allowed)
		resp, err := http.Get(testServer.URL + "/timeseries")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

		// POST to /range should fail (only GET allowed)
		resp2, err := http.Post(testServer.URL+"/range", "application/json", bytes.NewBufferString("{}"))
		require.NoError(t, err)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp2.StatusCode)
	})
}

func TestServerStartupAndShutdown(t *testing.T) {
	server := NewServer()

	// Test that the server is configured correctly
	assert.NotNil(t, server.Db)
	assert.NotNil(t, server.HttpServer)
	assert.Equal(t, ":8080", server.HttpServer.Addr)
	assert.NotNil(t, server.HttpServer.Handler)
}
