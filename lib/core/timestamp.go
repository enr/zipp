package core

import "time"

func Timestamp() string {
	const layout = "20060102150405"
	t := time.Now()
	ts := t.Local().Format(layout)
	return ts
}
