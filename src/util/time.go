package util

import "time"

func Now() string {
	now := time.Now()
	nowFormat := now.Format(time.RFC3339)

	return nowFormat
}
