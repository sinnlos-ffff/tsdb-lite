package database

import (
	"reflect"
	"testing"
)

func TestAddTimeSeries(t *testing.T) {
	db := NewDatabase()
	metric := "test_metric"
	tags := map[string]string{"tag1": "value1"}
	db.AddTimeSeries(metric, tags)

	key := generateKey(metric, tags)
	ts, ok := db.series[key]
	if !ok {
		t.Fatalf("TimeSeries not found for key: %s", key)
	}

	if ts.Metric != metric {
		t.Errorf("Expected metric %s, got %s", metric, ts.Metric)
	}

	if !reflect.DeepEqual(ts.Tags, tags) {
		t.Errorf("Expected tags %v, got %v", tags, ts.Tags)
	}
}
