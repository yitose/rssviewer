package feed

import "time"

func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	const timeFormat = "2006/01/02 15:04:05"
	return t.Format(timeFormat)
}

func FormatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	const format = "15:04"
	return t.Format(format)
}
