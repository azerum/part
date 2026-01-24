package utils_test

import (
	"errors"
	"testing"
	"time"

	"github.com/azerum/part/utils"
	. "github.com/onsi/gomega"
)

func zeroToFour() <-chan int {
	out := make(chan int)

	go func() {
		for i := range 5 {
			out <- i
		}

		close(out)
	}()

	return out
}

func drain[T any](ch <-chan T) []T {
	var out []T

	for v := range ch {
		out = append(out, v)
	}

	return out
}

func Test_runs_map_fn_with_given_concurrency_and_closes_once_input_closes(t *testing.T) {
	g := NewGomegaWithT(t)

	input := zeroToFour()

	mapFn := func(x int) *utils.ChanWithError[int] {
		out := utils.NewChanWithError[int](0)

		go func() {
			out.Channel <- -x
			out.Channel <- x
			out.CloseOk()
		}()

		return out
	}

	ch := utils.MapConcurrently(input, mapFn, 2)

	results := drain(ch.Channel)

	g.Expect(results).To(ConsistOf(0, 0, -1, 1, -2, 2, -3, 3, -4, 4))
	g.Expect(ch.Err).To(BeNil())
}

func Test_when_any_map_fn_fails_eventually_closes_output_and_returns_the_first_error(t *testing.T) {
	g := NewGomegaWithT(t)

	input := zeroToFour()

	mapFn := func(x int) *utils.ChanWithError[int] {
		out := utils.NewChanWithError[int](0)

		if x == 0 {
			time.Sleep(1 * time.Second)

			out.CloseWithError(errors.New("0"))
			return out
		}

		if x == 1 {
			time.Sleep(2 * time.Second)

			out.CloseWithError(errors.New("1"))
			return out
		}

		close(out.Channel)
		return out
	}

	ch := utils.MapConcurrently(input, mapFn, 2)

	// This indirectly checks that ch.Channel eventually closes, and waits
	// for it to close
	_ = drain(ch.Channel)

	// Error 0 happens earlier
	g.Expect(ch.Err).To(MatchError("0"))
}
