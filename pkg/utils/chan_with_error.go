package utils

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

func (c *ChanWithError[T]) CloseOk() {
	close(c.Channel)
}

func (c *ChanWithError[T]) CloseWithError(err error) {
	c.Err = err
	close(c.Channel)
}

func (c *ChanWithError[T]) Drain() ([]T, error) {
	values := make([]T, 0)

	for v := range c.Channel {
		values = append(values, v)
	}

	if c.Err != nil {
		return nil, c.Err
	}

	return values, nil
}
