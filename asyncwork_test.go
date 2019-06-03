package asyncwork

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GustavoKatel/asyncwork/interfaces"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
)

func setUpAsyncWorkTest(t *testing.T) (interfaces.AsyncWork, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	assert := assert.New(t)

	aw, err := New()
	assert.Nil(err)

	return aw, ctrl
}

func setDownAsyncWorkTest(ctrl *gomock.Controller) {
	ctrl.Finish()
}

func TestStartStop(t *testing.T) {
	aw, ctrl := setUpAsyncWorkTest(t)
	defer setDownAsyncWorkTest(ctrl)
	assert := assert.New(t)

	assert.Nil(aw.Start())
	assert.Nil(aw.Stop())
}

func TestRunOne(t *testing.T) {
	aw, ctrl := setUpAsyncWorkTest(t)
	defer setDownAsyncWorkTest(ctrl)
	assert := assert.New(t)

	assert.Nil(aw.Start())
	defer aw.Stop()

	waitCh := make(chan interface{}, 1)
	assert.Nil(aw.PostJob(func(ctx context.Context) error {
		waitCh <- nil
		return nil
	}))

	<-waitCh
}

func TestErrorOne(t *testing.T) {
	aw, ctrl := setUpAsyncWorkTest(t)
	defer setDownAsyncWorkTest(ctrl)
	assert := assert.New(t)

	assert.Nil(aw.Start())
	defer aw.Stop()

	errCh := make(chan error, 1)
	aw.ErrorChan(errCh)

	assert.Nil(aw.PostJob(func(ctx context.Context) error {
		return fmt.Errorf("Test")
	}))

	err := <-errCh
	assert.Equal(err.Error(), "Test")
}

func TestLen(t *testing.T) {
	aw, ctrl := setUpAsyncWorkTest(t)
	defer setDownAsyncWorkTest(ctrl)
	assert := assert.New(t)

	assert.Nil(aw.Start())
	defer aw.Stop()

	waitCh := make(chan interface{}, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	assert.Nil(aw.PostJob(func(ctx context.Context) error {
		<-waitCh
		wg.Done()
		return nil
	}))

	assert.Nil(aw.PostJob(func(ctx context.Context) error {
		<-waitCh
		wg.Done()
		return nil
	}))

	assert.Equal(2, aw.Len())
	waitCh <- nil
	waitCh <- nil
	wg.Wait()
	assert.Equal(0, aw.Len())
}

func TestRunTaggedOne(t *testing.T) {
	aw, ctrl := setUpAsyncWorkTest(t)
	defer setDownAsyncWorkTest(ctrl)
	assert := assert.New(t)

	assert.Nil(aw.Start())
	defer aw.Stop()

	var wg sync.WaitGroup
	wg.Add(1)
	assert.Nil(aw.PostTaggedJob(func(ctx context.Context) error {
		wg.Done()
		return nil
	}, "test"))

	wg.Wait()
}

func TestRunThrottledOne(t *testing.T) {
	aw, ctrl := setUpAsyncWorkTest(t)
	defer setDownAsyncWorkTest(ctrl)
	assert := assert.New(t)

	assert.Nil(aw.Start())
	defer aw.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	count := int64(0)

	assert.Nil(aw.PostThrottledJob(func(ctx context.Context) error {
		atomic.AddInt64(&count, 1)
		wg.Done()
		return nil
	}, 1*time.Second))

	assert.Nil(aw.PostThrottledJob(func(ctx context.Context) error {
		atomic.AddInt64(&count, 1)
		wg.Done()
		return nil
	}, 1*time.Second))

	wg.Wait()

	assert.Equal(int64(1), count)
}

func TestRunThrottledTwo(t *testing.T) {
	aw, ctrl := setUpAsyncWorkTest(t)
	defer setDownAsyncWorkTest(ctrl)
	assert := assert.New(t)

	assert.Nil(aw.Start())
	defer aw.Stop()

	var wg sync.WaitGroup
	wg.Add(2)

	count := int64(0)

	assert.Nil(aw.PostThrottledJob(func(ctx context.Context) error {
		atomic.AddInt64(&count, 1)
		wg.Done()
		return nil
	}, 500*time.Millisecond))

	<-time.After(501 * time.Millisecond)

	assert.Nil(aw.PostThrottledJob(func(ctx context.Context) error {
		atomic.AddInt64(&count, 1)
		wg.Done()
		return nil
	}, 500*time.Millisecond))

	wg.Wait()

	assert.Equal(int64(2), count)
}

func TestRunTaggedThrottledTwo(t *testing.T) {
	aw, ctrl := setUpAsyncWorkTest(t)
	defer setDownAsyncWorkTest(ctrl)
	assert := assert.New(t)

	assert.Nil(aw.Start())
	defer aw.Stop()

	var wg sync.WaitGroup
	wg.Add(3)

	count := int64(0)

	assert.Nil(aw.PostTaggedThrottledJob(func(ctx context.Context) error {
		atomic.AddInt64(&count, 1)
		wg.Done()
		return nil
	}, "500", 500*time.Millisecond))

	assert.Nil(aw.PostTaggedJob(func(ctx context.Context) error {
		<-time.After(501 * time.Millisecond)
		wg.Done()
		return nil
	}, "none"))

	assert.Nil(aw.PostTaggedThrottledJob(func(ctx context.Context) error {
		atomic.AddInt64(&count, 1)
		wg.Done()
		return nil
	}, "500", 500*time.Millisecond))

	wg.Wait()

	assert.Equal(int64(2), count)
}
