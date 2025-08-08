package database

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/sinnlos-ffff/tsdb-lite/metrics"
)

type Point struct {
	Timestamp int64
	Value     float64
}

const ChunkSize = 2048

type Chunk struct {
	Points    []Point
	Count     int
	Compacted bool
}

type TimeSeries struct {
	sync.RWMutex
	Metric string
	Tags   map[string]string
	Chunks []*Chunk
}

type Shard struct {
	sync.RWMutex
	Series map[string]*TimeSeries
}

func (s *Shard) CompactChunks() {
	s.Lock()
	defer s.Unlock()

	for _, ts := range s.Series {
		ts.Lock()
		defer ts.Unlock()

		for _, chunk := range ts.Chunks {
			if chunk.Compacted || chunk.Count < ChunkSize {
				continue
			}

			sort.Slice(chunk.Points[:ChunkSize], func(i, j int) bool {
				return chunk.Points[i].Timestamp < chunk.Points[j].Timestamp
			})

			chunk.Compacted = true
			metrics.CompactedChunksTotal.Inc()
		}
	}
}

func (s *Shard) RunCompactor(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			s.CompactChunks()
		}
	}()
}

type Database struct {
	Shards []*Shard
}

func NewDatabase() *Database {
	db := &Database{
		// TODO: Dynamically update shard length.
		Shards: make([]*Shard, 32),
	}
	for i := range db.Shards {
		db.Shards[i] = &Shard{
			Series: make(map[string]*TimeSeries),
		}
	}
	return db
}

func (db *Database) GetShard(key string) *Shard {
	shardIndex := xxhash.Sum64String(key) % uint64(len(db.Shards))
	return db.Shards[shardIndex]
}

func (db *Database) StartCompactors(interval time.Duration) {
	for i := range db.Shards {
		db.Shards[i].RunCompactor(interval)
	}
}

func (db *Database) AddTimeSeries(metric string, tags map[string]string) error {
	key := GenerateKey(metric, tags)
	shard := db.GetShard(key)

	// TODO: error handling
	if _, ok := shard.Series[key]; ok {
		return errors.New("time series already exists")
	}

	ts := TimeSeries{
		Metric: metric,
		Tags:   tags,
		Chunks: []*Chunk{{
			Points: make([]Point, 0, ChunkSize),
		}},
	}

	shard.Lock()
	defer shard.Unlock()

	shard.Series[key] = &ts

	return nil
}

func (db *Database) AddPoint(metric string, tags map[string]string, timestamp int64, value float64) error {
	start := time.Now()
	key := GenerateKey(metric, tags)
	shard := db.GetShard(key)

	shard.RLock()
	ts, ok := shard.Series[key]
	shard.RUnlock()

	if !ok {
		return errors.New("time series does not exist")
	}

	ts.Lock()
	defer ts.Unlock()

	if ts.Chunks[len(ts.Chunks)-1].Count == ChunkSize || ts.Chunks[len(ts.Chunks)-1].Compacted {
		ts.Chunks = append(ts.Chunks, &Chunk{
			Points:    make([]Point, 0, ChunkSize),
			Count:     0,
			Compacted: false,
		})
	}

	chunk := ts.Chunks[len(ts.Chunks)-1]
	chunk.Points = append(chunk.Points, Point{
		Timestamp: timestamp,
		Value:     value,
	})
	chunk.Count++

	metrics.IngestLatency.Observe(time.Since(start).Seconds())
	metrics.IngestTotal.Inc()

	return nil
}
