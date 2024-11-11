package util

import (
	"context"
)

type IO[T any] func(context.Context) (T, error)

func ComposeIo[T, U any](
	i IO[T],
	mapper func(T) U,
) IO[U] {
	return func(ctx context.Context) (u U, e error) {
		t, e := i(ctx)
		if nil != e {
			return u, e
		}
		u = mapper(t)
		return u, nil
	}
}

func ComposeIoErr[T, U any](
	i IO[T],
	mapper func(T) (U, error),
) IO[U] {
	return func(ctx context.Context) (u U, e error) {
		t, e := i(ctx)
		if nil != e {
			return u, e
		}
		return mapper(t)
	}
}
