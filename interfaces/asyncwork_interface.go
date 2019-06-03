package interfaces

import "time"

// AsyncWork provides a scheduling work channel
type AsyncWork interface {
	// Start starts the worker channel
	Start() error

	// Stop stops all running and scheduled jobs
	Stop() error

	// PostJob schedules a job execution
	PostJob(job JobFn) error

	// PostJob schedules a job execution and tags with "tag"
	PostTaggedJob(job JobFn, tag string) error

	// PostThrottledJob posts a job only and only if the time span of its last execution was greater than "duration"
	PostThrottledJob(job JobFn, delay time.Duration) error

	// PostTaggedThrottledJob same as PostThrottled but check only with jobs with the same tag
	PostTaggedThrottledJob(job JobFn, tag string, delay time.Duration) error

	// Len returns the number of jobs scheduled
	Len() int

	// ErrorChan registers an error emitting channel
	ErrorChan(ch chan error)
}
