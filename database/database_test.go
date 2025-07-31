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
	assert.Len(t, ts.Points, 1, "Expected 1 point in TimeSeries")
	assert.Equal(t, timestamp, ts.Points[0].Timestamp)
	assert.Equal(t, value, ts.Points[0].Value)
}
