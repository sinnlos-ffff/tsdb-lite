package storage

// Example put request body from OpenTSDB
// {
//     "metric": "sys.cpu.nice",
//     "timestamp": 1346846400,
//     "value": 18,
//     "tags": {
//        "host": "web01",
//        "dc": "lga"
//     }
// }

type DataPoint struct {
	Timestamp int64
	Value     float64
}

type TimeSeries struct {
	Metric string
	Tags   map[string]string
	Points []DataPoint // TODO: sort by timestamp when possible
}
