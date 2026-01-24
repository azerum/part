package utils

import (
	"context"
	"sync"
)

type ChanWithError[T any] struct {
	Channel chan T
	Err     error
}

func NewChanWithError[T any](size int) *ChanWithError[T] {
	return &ChanWithError[T]{
		Channel: make(chan T, size),
		Err:     nil,
	}
}

type MapFn[I any, O any] func(input I) *ChanWithError[O]

func MapConcurrently[I any, O any](
	input <-chan I,
	mapFn MapFn[I, O],
	concurrency int,
) *ChanWithError[O] {
	outputs := make([]*ChanWithError[O], 0, concurrency)

	for range concurrency {
		ch := Map(input, mapFn)
		outputs = append(outputs, ch)
	}

	return merge(outputs)
}

func Map[I any, O any](
	input <-chan I,
	mapFn MapFn[I, O],
) *ChanWithError[O] {
	out := NewChanWithError[O](0)

	go func() {
		defer close(out.Channel)

		for v := range input {
			ch := mapFn(v)

			for o := range ch.Channel {
				out.Channel <- o
			}

			if ch.Err != nil {
				out.Err = ch.Err
				return
			}
		}
	}()

	return out
}

func merge[T any](channels []*ChanWithError[T]) *ChanWithError[T] {
	out := NewChanWithError[T](len(channels))

	// Done when all workers exit
	var wg sync.WaitGroup

	// Once one worker observes error, it signals other workers to exit
	// via this context. The context also stores the first observed error
	ctx, cancel := context.WithCancelCause(context.Background())

	worker := func(ch *ChanWithError[T]) {
		for {
			select {
			case v, ok := <-ch.Channel:
				if ok {
					out.Channel <- v
					continue
				}

				// ch.Channel closed. Either cleanly or with error
				// If with error, we should notify other workers

				if ch.Err != nil {
					cancel(ch.Err)
				}

				return

			case <-ctx.Done():
				// One of the passed channels has closed with error. Stop
				// reading from the remaining channels as soon as possible
				return
			}
		}
	}

	for _, c := range channels {
		wg.Go(func() {
			worker(c)
		})
	}

	go func() {
		wg.Wait()

		if err := context.Cause(ctx); err != nil {
			out.Err = err
		}

		close(out.Channel)
	}()

	return out
}
