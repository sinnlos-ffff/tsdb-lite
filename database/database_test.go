package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddTimeSeries(t *testing.T) {
	db := NewDatabase()
	metric := "test_metric"
	tags := map[string]string{"tag1": "value1"}
	db.AddTimeSeries(metric, tags)

	key := GenerateKey(metric, tags)
	ts, ok := db.GetShard(key).Series[key]

	assert.True(t, ok, "TimeSeries not found for key: %s", key)
	assert.Equal(t, metric, ts.Metric)
	assert.Equal(t, tags, ts.Tags)
}

func TestAddPoint(t *testing.T) {
	db := NewDatabase()
	metric := "test_metric"
	tags := map[string]string{"tag1": "value1"}
	db.AddTimeSeries(metric, tags)

	timestamp := int64(123456789)
	value := 10.5
	db.AddPoint(metric, tags, timestamp, value)

	key := GenerateKey(metric, tags)
	ts, ok := db.GetShard(key).Series[key]

	assert.True(t, ok, "TimeSeries not found for key: %s", key)
	assert.Len(t, ts.Chunks, 1, "Expected 1 Chunk")

	chunk := ts.Chunks[0]
	assert.Len(t, chunk.Points, 1, "Expected 1 Point")
	assert.Equal(t, timestamp, chunk.Points[0].Timestamp)
	assert.Equal(t, value, chunk.Points[0].Value)
}
