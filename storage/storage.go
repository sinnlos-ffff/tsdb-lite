package storage

type DataPoint struct {
    Timestamp int64
    Value     float64
}

type TimeSeries struct {
    Metric string
    Points []DataPoint // kept sorted by timestamp
}
