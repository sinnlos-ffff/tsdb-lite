package database

import (
	"errors"
	"sync"

	"github.com/cespare/xxhash/v2"
)

type Point struct {
	Timestamp int64
	Value     float64
}

type TimeSeries struct {
	sync.Mutex
	Metric string
	Tags   map[string]string
	Points []Point // TODO: sort by timestamp when possible
}

type Shard struct {
	sync.RWMutex
	Series map[string]*TimeSeries
}

type Database struct {
	Shards []*Shard
}

func NewDatabase() *Database {
	db := &Database{
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
		Points: make([]Point, 0),
	}

	shard.Lock()
	defer shard.Unlock()

	shard.Series[key] = &ts

	return nil
}

func (db *Database) AddPoint(metric string, tags map[string]string, timestamp int64, value float64) error {
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

	ts.Points = append(ts.Points, Point{
		Timestamp: timestamp,
		Value:     value,
	})

	return nil
}
