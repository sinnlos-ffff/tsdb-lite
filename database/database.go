package database

import (
	"errors"
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
	Series map[string]*TimeSeries
}

func NewDatabase() *Database {
	return &Database{
		Series: make(map[string]*TimeSeries),
	}
}

func (db *Database) AddTimeSeries(metric string, tags map[string]string) error {
	db.Lock()
	defer db.Unlock()

	key := GenerateKey(metric, tags)

	// TODO: error handling
	if _, ok := db.Series[key]; ok {
		return errors.New("time series already exists")
	}

	ts := TimeSeries{
		Metric: metric,
		Tags:   tags,
		Points: make([]Point, 0),
	}

	db.Series[key] = &ts

	return nil
}
