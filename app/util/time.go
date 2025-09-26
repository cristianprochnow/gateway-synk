package util

import "time"

func Now() string {
	now := time.Now()
	nowFormat := now.Format(time.RFC3339)

	return nowFormat
}

func ToTimeBR(date string) string {
	originalDateStr := date
	inputLayout := "2006-01-02 15:04:05"
	outputLayout := "02/01/2006 15:04:05"

	t, err := time.Parse(inputLayout, originalDateStr)

	if err != nil {
		return ""
	}

	formattedDateStr := t.Format(outputLayout)

	return formattedDateStr
}
