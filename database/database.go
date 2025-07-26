package database

import (
	"sync"
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

type Database struct {
	sync.RWMutex
	series map[string]*TimeSeries
}

func NewDatabase() *Database {
	return &Database{
		series: make(map[string]*TimeSeries),
	}
}

func (db *Database) AddTimeSeries(metric string, tags map[string]string) {
	db.Lock()
	defer db.Unlock()

	key := generateKey(metric, tags)

	// TODO: error handling
	if _, ok := db.series[key]; ok {
		return
	}

	ts := TimeSeries{
		Metric: metric,
		Tags:   tags,
		Points: make([]Point, 0),
	}

	db.series[key] = &ts
}
