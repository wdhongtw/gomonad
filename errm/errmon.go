package errm

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
)

type Monad[T any] struct {
	value *T
	err   error
}

func Return[T any](value T) Monad[T] {
	return Some(value)
}

func Some[T any](value T) Monad[T] {
	return Monad[T]{
		value: &value,
		err:   nil,
	}
}

func Err[T any](err error) Monad[T] {
	return Monad[T]{
		value: nil,
		err:   err,
	}
}

func Bind[T, U any](
	mt Monad[T], action func(t T) Monad[U],
) Monad[U] {
	if mt.err != nil {
		return Monad[U]{err: mt.err}
	}
	return action(*mt.value)
}

func WithDefault[T any](
	mt Monad[T], action func() Monad[T],
) Monad[T] {
	if mt.err != nil {
		return action()
	}
	return mt
}

func AndThen[T, U any](
	mt Monad[T], action func(t T) Monad[U],
) Monad[U] {
	return Bind(mt, action)
}

func OrElse[T any](
	mt Monad[T], action func() Monad[T],
) Monad[T] {
	return WithDefault(mt, action)
}

func FromValue[T any](t T) func() Monad[T] {
	return func() Monad[T] {
		return Monad[T]{value: &t}
	}
}

func Build[T any](action func() (T, error)) func() Monad[T] {
	t, err := action()
	return func() Monad[T] {
		return Monad[T]{value: &t, err: err}
	}
}

func Fish[T, U, V any](
	actionA func(t T) Monad[U],
	actionB func(u U) Monad[V],
) func(t T) Monad[V] {
	return func(t T) Monad[V] {
		mu := actionA(t)
		if mu.err != nil {
			return Monad[V]{value: nil, err: mu.err}
		}
		return actionB(*mu.value)
	}
}

func Join[T any](mmt Monad[Monad[T]]) Monad[T] {
	if mmt.err != nil {
		return Monad[T]{err: mmt.err}
	}
	return *mmt.value
}

func Raise[T, U any](action func(T) Monad[U]) func(Monad[T]) Monad[Monad[U]] {
	return func(mt Monad[T]) Monad[Monad[U]] {
		if mt.err != nil {
			return Monad[Monad[U]]{err: mt.err}
		}
		mu := action(*mt.value)
		return Monad[Monad[U]]{value: &mu}
	}
}

func Wrap[T, U any](action func(T) (U, error)) func(T) Monad[U] {
	return func(t T) Monad[U] {
		u, err := action(t)
		if err != nil {
			if decorateError {
				err = fmt.Errorf("%v(%#v): %w", getFunctionName(action), t, err)
			}
			return Monad[U]{value: nil, err: err}
		}
		return Monad[U]{value: &u, err: nil}
	}
}

func Ensure[T, U any](action func(T) U) func(T) Monad[U] {
	return func(t T) Monad[U] {
		mu := Wrap(func(T) (U, error) {
			return action(t), nil
		})(t)
		return mu
	}
}

func Unpack[T any](t Monad[T]) (T, error) {
	if t.value != nil {
		return *t.value, t.err
	}
	var dummy T
	return dummy, t.err
}

func Transform[T, U any](t Monad[T], action func(t T) (U, error)) Monad[U] {
	return Bind(t, Wrap(action))
}

func getFunctionName(function interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
}

func WithCtx[T, U any](
	ctx context.Context, action func(ctx context.Context, t T) (U, error),
) func(t T) (U, error) {
	return func(t T) (U, error) {
		return action(ctx, t)
	}
}

var decorateError bool = true

func DisableErrorDecoration() {
	decorateError = false
}

func EnableErrorDecoration() {
	decorateError = true
}
