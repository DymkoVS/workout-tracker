package model

import "time"

type AnalyticsPoint struct {
	Date  time.Time
	Value float64
}

type FrequencyPoint struct {
	Week  string
	Count int
}
