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
	ts, ok := db.Series[key]

	assert.True(t, ok, "TimeSeries not found for key: %s", key)
	assert.Equal(t, metric, ts.Metric)
	assert.Equal(t, tags, ts.Tags)
}