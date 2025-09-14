package system

import "time"

// SystemClock は実際のシステム時刻を返すClock実装です。
type SystemClock struct{}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

func (c *SystemClock) Now() time.Time {
	return time.Now()
}