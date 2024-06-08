package utils

import (
	"time"
)

func GetLocalTimeZoneOffset() int {
	_, offset := time.Now().Zone()
	return offset
}