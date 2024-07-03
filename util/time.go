package util

import (
	"time"
)

func Millisecond() uint32 {
	ms := time.Now().UnixNano() / 1e6
	return uint32(ms)
}

var TimestampFormat = "2006-01-02 15:04:05"

func FormatTime(tm *time.Time) string {
	if tm == nil {
		return ""
	}
	return tm.Format(TimestampFormat)
}

var TimestampFormat2 = "2006-01-02 15:04:05.000"

func FormatTimeMillisecond(tm *time.Time) string {
	if tm == nil {
		return ""
	}
	return tm.Format(TimestampFormat2)
}
