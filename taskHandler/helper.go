package main

import (
	"fmt"
	"log"
	"time"
)

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}

func getNow() time.Time {
	nowBase := time.Now()
	now, err := time.Parse(config.Layout, fmt.Sprintf("%02d:%02d", nowBase.Hour(), nowBase.Minute()))
	if err != nil {
		println(err.Error())
		log.Fatal("Failed to assemble current time defined by layout (14:05)")
	}
	return now
}
