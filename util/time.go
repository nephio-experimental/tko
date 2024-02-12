package util

import (
	"time"
)

func SecondsToDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}
