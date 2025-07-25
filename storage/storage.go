package storage

type Point struct {
	Timestamp int64
	Value     float64
}

type TimeSeries struct {
	Metric string
	Tags   map[string]string
	Points []Point // TODO: sort by timestamp when possible
}
