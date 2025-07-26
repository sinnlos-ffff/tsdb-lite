package database

import (
	"bytes"
	"sort"
)

// TODO: Add test
func generateKey(metric string, tags map[string]string) string {
	if len(tags) == 0 {
		return metric
	}

	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b bytes.Buffer
	b.WriteString(metric)
	b.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(tags[k])
	}
	b.WriteString("}")

	return b.String()
}
