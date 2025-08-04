package database

import "errors"

func (db *Database) GetRange(metric string, tags map[string]string, start, end int64) ([]Point, error) {
	key := GenerateKey(metric, tags)
	shard := db.GetShard(key)

	shard.RLock()
	defer shard.RUnlock()

	timeSeries, exists := shard.Series[key]
	if !exists {
		return nil, errors.New("time series not found")
	}

	var result []Point
	for _, chunk := range timeSeries.Chunks {
		for _, point := range chunk.Points {
			if point.Timestamp >= start && point.Timestamp <= end {
				result = append(result, point)
			}
		}
	}

	return result, nil
}
