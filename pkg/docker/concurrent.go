package docker

import (
	"context"
	"fmt"
	"sync"
)

type ConcurrentImageProcessor struct {
	maxRoutines int
}

func NewConcurrentImageProcessor(maxRoutines int) *ConcurrentImageProcessor {
	return &ConcurrentImageProcessor{maxRoutines: maxRoutines}
}

type ImageProcessor func(ctx context.Context, image string) error

func (c *ConcurrentImageProcessor) Process(ctx context.Context, images []string, process ImageProcessor) error {
	contextWithCancel, abortRemainingJobs := context.WithCancel(ctx)
	workers := min(len(images), c.maxRoutines)
	workers = max(workers, 1)
	jobsChan := make(chan job)
	wg := &sync.WaitGroup{}
	errors := make(chan error)
	doneChan := make(chan struct{})

	for i := 0; i < workers; i++ {
		w := &worker{
			jobs:        jobsChan,
			process:     process,
			waitGroup:   wg,
			errorReturn: errors,
		}

		wg.Add(1)
		go w.start()
	}

	f := &feeder{
		jobs:            jobsChan,
		images:          images,
		workersWaiGroup: wg,
		done:            doneChan,
	}
	go f.feed(contextWithCancel)

	var firstError error

loop:
	for {
		select {
		case <-doneChan:
			break loop
		case err := <-errors:
			if firstError == nil {
				firstError = err
				abortRemainingJobs()
			}
		}
	}

	// This is not necessary because if we get here all workers are done, regardless if all the jobs
	// were run or not, so canceling the context is not necessary since there is nothing else using it
	// This is just to avoid a lint warning for possible context leeking
	abortRemainingJobs()

	if firstError != nil {
		return fmt.Errorf("image processor worker failed, rest of jobs were aborted: %v", firstError)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type job struct {
	ctx   context.Context
	image string
}

type feeder struct {
	jobs            chan<- job
	images          []string
	done            chan<- struct{}
	workersWaiGroup *sync.WaitGroup
}

func (f *feeder) feed(ctx context.Context) {
	defer func() {
		close(f.jobs)
		f.workersWaiGroup.Wait()
		close(f.done)
	}()

	for _, i := range f.images {
		select {
		case <-ctx.Done():
			return
		default:
			j := job{
				ctx:   ctx,
				image: i,
			}
			f.jobs <- j
		}
	}
}

type worker struct {
	jobs        <-chan job
	process     ImageProcessor
	waitGroup   *sync.WaitGroup
	errorReturn chan<- error
}

func (w *worker) start() {
	defer w.waitGroup.Done()
	for j := range w.jobs {
		if err := w.process(j.ctx, j.image); err != nil {
			w.errorReturn <- err
		}
	}
}
