package asyncwork

import (
	"context"
	"sync"
	"time"

	"github.com/GustavoKatel/asyncwork/interfaces"
	"github.com/GustavoKatel/syncevent"
	"github.com/phf/go-queue/queue"
)

var _ interfaces.AsyncWork = &asyncWorkImpl{}

type asyncWorkImpl struct {
	queue      *queue.Queue
	queueMutex *sync.RWMutex

	ctx       context.Context
	ctxCancel context.CancelFunc

	event syncevent.SyncEvent

	errorChs      []chan error
	errorChsMutex *sync.RWMutex

	lastExecutionMap      map[string]time.Time
	lastExecutionMapMutex *sync.RWMutex

	// Maps tag to the last recorded call
	throttleLastMap      map[string]*jobImpl
	throttleLastMapMutex *sync.RWMutex
}

// New creates a new working channel
func New() (interfaces.AsyncWork, error) {
	return NewWithContext(context.Background())
}

// NewWithContext creates a new working channel
func NewWithContext(ctx context.Context) (interfaces.AsyncWork, error) {
	ctx, cancel := context.WithCancel(ctx)

	aw := &asyncWorkImpl{
		queue:      queue.New(),
		queueMutex: &sync.RWMutex{},

		ctx:       ctx,
		ctxCancel: cancel,

		event: syncevent.NewSyncEvent(false),

		errorChs:      []chan error{},
		errorChsMutex: &sync.RWMutex{},

		lastExecutionMap:      map[string]time.Time{},
		lastExecutionMapMutex: &sync.RWMutex{},

		throttleLastMap:      map[string]*jobImpl{},
		throttleLastMapMutex: &sync.RWMutex{},
	}

	return aw, nil
}

func (aw *asyncWorkImpl) Start() error {
	go aw.background()
	return nil
}

func (aw *asyncWorkImpl) Stop() error {
	aw.ctxCancel()
	aw.event.Set()
	return nil
}

func (aw *asyncWorkImpl) ErrorChan(ch chan error) {
	aw.errorChsMutex.RLock()
	defer aw.errorChsMutex.RUnlock()

	aw.errorChs = append(aw.errorChs, ch)
}

func (aw *asyncWorkImpl) emitError(err error) {
	aw.errorChsMutex.RLock()
	defer aw.errorChsMutex.RUnlock()

	for _, ch := range aw.errorChs {
		ch <- err
	}
}

func (aw *asyncWorkImpl) background() {
	for aw.ctx.Err() == nil {
		aw.event.Wait()
		if aw.ctx.Err() != nil {
			break
		}

		aw.queueMutex.RLock()
		if aw.queue.Len() == 0 {
			aw.event.Reset()
			aw.queueMutex.RUnlock()
			continue
		}
		aw.queueMutex.RUnlock()

		if err := aw.fetchAndRun(); err != nil {
			aw.emitError(err)
		}
	}
}

func (aw *asyncWorkImpl) waitAndSchedule(job *jobImpl, delay time.Duration) {
	aw.throttleLastMapMutex.Lock()

	_, prs := aw.throttleLastMap[job.tag]
	aw.throttleLastMap[job.tag] = job

	aw.throttleLastMapMutex.Unlock()

	if prs {
		return
	}

	tag := job.tag

	go func() {
		select {
		case <-time.After(delay):
			aw.throttleLastMapMutex.Lock()
			defer aw.throttleLastMapMutex.Unlock()

			lastJob, prs := aw.throttleLastMap[tag]
			if !prs {
				return
			}

			delete(aw.throttleLastMap, tag)

			aw.PostTaggedThrottledJob(lastJob.jobFn, lastJob.tag, lastJob.delay)
		case <-aw.ctx.Done():
			return
		}
	}()
}

func (aw *asyncWorkImpl) fetchAndRun() error {
	aw.queueMutex.Lock()
	jobI := aw.queue.PopFront()
	aw.queueMutex.Unlock()

	job := jobI.(*jobImpl)

	aw.lastExecutionMapMutex.Lock()
	defer aw.lastExecutionMapMutex.Unlock()

	lastExecution, prs := aw.lastExecutionMap[job.tag]
	if prs && time.Now().Sub(lastExecution) < job.delay {
		aw.waitAndSchedule(job, job.delay-time.Now().Sub(lastExecution))
		return nil
	}

	aw.lastExecutionMap[job.tag] = time.Now()

	return job.jobFn(aw.ctx)
}

func (aw *asyncWorkImpl) PostJob(job interfaces.JobFn) error {
	return aw.PostTaggedJob(job, "")
}

func (aw *asyncWorkImpl) PostTaggedJob(job interfaces.JobFn, tag string) error {
	return aw.PostTaggedThrottledJob(job, tag, 0)
}

func (aw *asyncWorkImpl) PostThrottledJob(job interfaces.JobFn, delay time.Duration) error {
	return aw.PostTaggedThrottledJob(job, "", delay)
}

func (aw *asyncWorkImpl) PostTaggedThrottledJob(job interfaces.JobFn, tag string, delay time.Duration) error {
	aw.queueMutex.Lock()
	defer aw.queueMutex.Unlock()

	jobData := jobImpl{
		jobFn: job,
		tag:   tag,
		delay: delay,
	}

	aw.queue.PushBack(&jobData)

	aw.event.Set()
	return nil
}

func (aw *asyncWorkImpl) Len() int {
	aw.queueMutex.RLock()
	defer aw.queueMutex.RUnlock()

	return aw.queue.Len()
}
