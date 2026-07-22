package entity

import (
	"strconv"
	"time"
)

func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "0 m"
	}
	switch {
	case d%(24*time.Hour) == 0:
		return strconv.Itoa(int(d/(24*time.Hour))) + " d"
	case d%time.Hour == 0:
		return strconv.Itoa(int(d/time.Hour)) + " h"
	case d < time.Minute:
		return strconv.Itoa(int(d/time.Second)) + " s"
	default:
		return strconv.Itoa(int(d/time.Minute)) + " m"
	}
}
