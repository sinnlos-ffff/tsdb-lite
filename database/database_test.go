package database

import (
	"math/rand"
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

func TestCompactChunks(t *testing.T) {
	shard := &Shard{
		Series: make(map[string]*TimeSeries),
	}

	metric := "test_metric"
	tags := map[string]string{"tag1": "value1"}
	key := GenerateKey(metric, tags)

	ts := &TimeSeries{
		Metric: metric,
		Tags:   tags,
		Chunks: []*Chunk{{
			Points: make([]Point, ChunkSize),
			Count:  ChunkSize,
		}},
	}

	// Add unsorted points to the chunk
	ts.Chunks[0].Points[0] = Point{Timestamp: 3, Value: 3.0}
	ts.Chunks[0].Points[1] = Point{Timestamp: 1, Value: 1.0}
	ts.Chunks[0].Points[2] = Point{Timestamp: 2, Value: 2.0}

	// Fill the rest of the chunk with dummy data to reach ChunkSize
	for i := 0; i < ChunkSize; i++ {
		ts.Chunks[0].Points[i] = Point{Timestamp: rand.Int63(), Value: float64(i)}
	}

	shard.Series[key] = ts

	shard.CompactChunks()

	// Assert chunk is compacted
	assert.True(t, ts.Chunks[0].Compacted)
	for i := 1; i < ChunkSize; i++ {
		assert.True(t, ts.Chunks[0].Points[i].Timestamp > ts.Chunks[0].Points[i-1].Timestamp)
	}
}
