package stamps

import "time"

type ErrorDetailsStamp struct {
	ErrorMessage string
	FailedAt     time.Time
	RetryCount   uint
}
