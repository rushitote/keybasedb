package main

import (
	"encoding/binary"
	"time"
)

func AddTimestampToValue(value string) string {
	now := make([]byte, 8)
	binary.BigEndian.PutUint64(now, uint64(time.Now().UnixNano()))
	return string(now) + value
}

func GetTimestampFromValue(value string) string {
	if value == "" {
		return ""
	}
	return value[:8]
}

func GetValueTextFromValue(value string) string {
	if value == "" {
		return ""
	}
	return value[8:]
}
