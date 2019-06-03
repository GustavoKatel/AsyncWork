# AsyncWork

[![GoDoc](https://godoc.org/github.com/GustavoKatel/asyncwork?status.svg)](https://godoc.org/github.com/GustavoKatel/asyncwork)

AsyncWork schedules and throttles function execution in golang

## Examples

### Scheduling
```go
var wg sync.WaitGroup
wg.Add(2)

worker, err := asyncwork.New()
if err != nil {
	panic(err)
}
worker.Start()
defer worker.Stop()

worker.PostJob(func(ctx context.Context) error {
	// Long operation 1
	log.Printf("Operation1")
	wg.Done()
	return nil
})

worker.PostJob(func(ctx context.Context) error {
	// Long operation 2
	log.Printf("Operation2")
	wg.Done()
	return nil
})

log.Printf("Pending: %v", worker.Len())
wg.Wait()
log.Printf("Pending: %v", worker.Len())
```

### Output
```
Pending: 2
Operation1
Operation2
Pending: 0
```

### Throttle
```go
var wg sync.WaitGroup
wg.Add(2)

worker, err := asyncwork.New()
if err != nil {
	panic(err)
}
worker.Start()
defer worker.Stop()

worker.PostJob(func(ctx context.Context) error {
	// Long operation 1
	log.Printf("Operation1")
	wg.Done()
	return nil
})

worker.PostThrottledJob(func(ctx context.Context) error {
	// Long operation 2 is not executed due to throttle
	log.Printf("Operation2")
	return nil
}, 500*time.Millisecond)

<-time.After(600 * time.Millisecond)

worker.PostThrottledJob(func(ctx context.Context) error {
	// Long operation 3
	log.Printf("Operation3")
	wg.Done()
	return nil
}, 500*time.Millisecond)

wg.Wait()
log.Printf("Pending: %v", worker.Len())
```

### Output
```
Operation1
Operation3
Pending: 0
```