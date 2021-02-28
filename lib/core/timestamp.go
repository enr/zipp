package core

import "time"

// Timestamp return a string representation of the timestamp in format "yyyymmddhhmmss"
func Timestamp() string {
	const layout = "20060102150405"
	t := time.Now()
	ts := t.Local().Format(layout)
	return ts
}
