package server

import (
	"context"
	"testing"
	"time"

	"github.com/sinnlos-ffff/tsdb-lite/database"
	pb "github.com/sinnlos-ffff/tsdb-lite/proto"
	"github.com/stretchr/testify/assert"
)

func TestCreateTimeSeries(t *testing.T) {
	db := database.NewDatabase()
	server := &Server{Db: db}

	metric := "test_metric"
	tags := map[string]string{"tag1": "value1"}
	req := &pb.CreateTimeSeriesRequest{
		Metric: metric,
		Tags:   tags,
	}

	_, err := server.CreateTimeSeries(context.Background(), req)
	assert.NoError(t, err)

	key := database.GenerateKey(metric, tags)
	ts, ok := db.GetShard(key).Series[key]

	assert.True(t, ok, "TimeSeries not found for key: %s", key)
	assert.Equal(t, metric, ts.Metric)
	assert.Equal(t, tags, ts.Tags)

	// Re-creating a time series with the same metric and tags returns an error
	_, err = server.CreateTimeSeries(context.Background(), req)
	assert.Error(t, err)
}

func TestAddPoint(t *testing.T) {
	db := database.NewDatabase()
	server := &Server{Db: db}

	metric := "test_metric"
	tags := map[string]string{"tag1": "value1"}
	timestamp := time.Now().Unix()
	value := 1.23

	// Add a time series to the database
	err := db.AddTimeSeries(metric, tags)
	assert.NoError(t, err)

	req := &pb.AddPointRequest{
		Metric:    metric,
		Timestamp: timestamp,
		Value:     value,
		Tags:      tags,
	}

	_, err = server.AddPoint(context.Background(), req)
	assert.NoError(t, err)

	// Check if the point was added to the time series
	key := database.GenerateKey(metric, tags)
	ts, ok := db.GetShard(key).Series[key]
	assert.True(t, ok)
	assert.Equal(t, 1, len(ts.Chunks[0].Points))
	assert.Equal(t, timestamp, ts.Chunks[0].Points[0].Timestamp)
	assert.Equal(t, value, ts.Chunks[0].Points[0].Value)
}

func TestGetRange(t *testing.T) {
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

	req := &pb.GetRangeRequest{
		Metric: metric,
		Tags:   tags,
		Start:  0,
		End:    4000,
	}

	resp, err := server.GetRange(context.Background(), req)
	assert.NoError(t, err)
	assert.Len(t, resp.Points, 3)

	// Test range that includes only the first two points
	req = &pb.GetRangeRequest{
		Metric: metric,
		Tags:   tags,
		Start:  500,
		End:    2500,
	}

	resp, err = server.GetRange(context.Background(), req)
	assert.NoError(t, err)
	assert.Len(t, resp.Points, 2)
	assert.Equal(t, value1, resp.Points[0].Value)
	assert.Equal(t, value2, resp.Points[1].Value)
}
