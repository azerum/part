package main

import (
	"context"
	"sync"
)

func mapConcurrently[I any, O any](
	input <-chan I,
	mapFn func(I, context.Context) ([]O, error),
	concurrency int,
	ctx context.Context,
	cancel context.CancelCauseFunc,
) <-chan O {
	worker := func() <-chan O {
		out := make(chan O, 1)

		go func() {
			defer close(out)

			for {
				select {
				case <-ctx.Done():
					return

				case i, ok := <-input:
					if !ok {
						return
					}

					results, err := mapFn(i, ctx)

					if err != nil {
						cancel(err)
						return
					}

					for _, r := range results {
						out <- r
					}
				}
			}
		}()

		return out
	}

	outputs := make([]<-chan O, 0)

	for range concurrency {
		out := worker()
		outputs = append(outputs, out)
	}

	return merge(outputs)
}

func merge[T any](channels []<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup

	for _, c := range channels {
		wg.Go(func() {
			for v := range c {
				out <- v
			}
		})
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
