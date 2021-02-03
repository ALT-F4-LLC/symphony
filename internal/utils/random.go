package utils

import (
	"math/rand"
	"time"
)

func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func RandomTimeout(maxMilliseconds int) time.Duration {
	offset := RandomInt(0, maxMilliseconds)

	return time.Duration(offset) * time.Millisecond
}
