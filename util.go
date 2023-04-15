package main

import (
	"fmt"
	"time"
)

// TODO: need to use better serialization method

func AddTimestampToValue(value string) string {
	if value == "" {
		return ""
	}
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	return timestamp + value
}

func GetTimestampFromValue(value string) string {
	if value == "" {
		return ""
	}
	return value[:10]
}

func GetValueTextFromValue(value string) string {
	if value == "" {
		return ""
	}
	return value[10:]
}

type HashRange struct {
	Low string
	High string
}