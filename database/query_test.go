package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRange(t *testing.T) {
	db := NewDatabase()
	metric := "test_metric"
	tags := map[string]string{"tag1": "value1"}
	db.AddTimeSeries(metric, tags)

	timestamp1 := int64(1000)
	value1 := 10.5
	db.AddPoint(metric, tags, timestamp1, value1)

	timestamp2 := int64(2000)
	value2 := 20.5
	db.AddPoint(metric, tags, timestamp2, value2)

	timestamp3 := int64(3000)
	value3 := 30.5
	db.AddPoint(metric, tags, timestamp3, value3)

	// Test range that includes all points
	points, err := db.GetRange(metric, tags, 0, 4000)
	assert.NoError(t, err)
	assert.Len(t, points, 3)

	// Test range that includes only the first two points
	points, err = db.GetRange(metric, tags, 500, 2500)
	assert.NoError(t, err)
	assert.Len(t, points, 2)
	assert.Equal(t, value1, points[0].Value)
	assert.Equal(t, value2, points[1].Value)

	// Test range that includes no points
	points, err = db.GetRange(metric, tags, 4000, 5000)
	assert.NoError(t, err)
	assert.Len(t, points, 0)

	// Test range with invalid metric
	invalidMetric := "invalid_metric"
	points, err = db.GetRange(invalidMetric, tags, 0, 4000)
	assert.Error(t, err)
	assert.Nil(t, points)
}
