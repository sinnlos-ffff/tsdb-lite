package storage

type DataPoint struct {
	Timestamp int64
	Value     float64
}

type TimeSeries struct {
	Metric string
	Tags   map[string]string
	Points []DataPoint // TODO: sort by timestamp when possible
}
