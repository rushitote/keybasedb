package main

import "time"

func AddTimestampToValue(value string) string {
	now, err := time.Now().MarshalText()
	if err != nil {
		panic(err)
	}
	return string(now) + value
}

func GetTimestampFromValue(value string) string {
	if value == "" {
		return ""
	}
	return value[:35]
}

func GetValueTextFromValue(value string) string {
	if value == "" {
		return ""
	}
	return value[35:]
}
