package interfaces

import "context"

// JobFn job interface
type JobFn func(ctx context.Context) error
