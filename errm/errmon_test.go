package errm

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
)

func TestMonad(t *testing.T) {
	var errorMultiply error
	var errorAdd error
	multiplyTwo := func(in int) (int, error) {
		return in * 2, errorMultiply
	}
	addThree := func(in int) (int, error) {
		return in + 3, errorAdd
	}

	t.Run("BindStyleSuccessChaining", func(t *testing.T) {
		errorMultiply = nil
		errorAdd = nil

		result, err := Unpack(Bind(Bind(
			Return(3),
			Wrap(multiplyTwo)),
			Wrap(addThree),
		))
		if err != nil {
			t.Errorf("expect no error, got: %v", err)
		}
		if result != 9 {
			t.Errorf("expect 9, got: %v", result)
		}
	})
	t.Run("BindStyleAbortOnError", func(t *testing.T) {
		errorMultiply = fmt.Errorf("some-error")
		errorAdd = nil

		result, err := Unpack(Bind(Bind(
			Return(3),
			Wrap(multiplyTwo)),
			Wrap(addThree),
		))
		t.Logf("got error: %v", err)
		if !errors.Is(err, errorMultiply) {
			t.Errorf("expect multiply error, got: %v", err)
		}
		if result != 0 {
			t.Errorf("expect 0, got: %v", result)
		}
	})
	t.Run("FishStyleSuccessChaining", func(t *testing.T) {
		errorMultiply = nil
		errorAdd = nil

		result, err := Unpack(Fish(
			Wrap(multiplyTwo),
			Wrap(addThree),
		)(3))
		if err != nil {
			t.Errorf("expect no error, got: %v", err)
		}
		if result != 9 {
			t.Errorf("expect 9, got: %v", result)
		}
	})
	t.Run("FishStyleAbortOnError", func(t *testing.T) {
		errorMultiply = fmt.Errorf("some-error")
		errorAdd = nil

		result, err := Unpack(Fish(
			Wrap(multiplyTwo),
			Wrap(addThree),
		)(3))
		t.Logf("got error: %v", err)
		if !errors.Is(err, errorMultiply) {
			t.Errorf("expect multiply error, got: %v", err)
		}
		if result != 0 {
			t.Errorf("expect 0, got: %v", result)
		}
	})
	t.Run("JoinStyleSuccessChaining", func(t *testing.T) {
		errorMultiply = nil
		errorAdd = nil

		result, err := Unpack(Join(
			Raise(Wrap(addThree))(Wrap(multiplyTwo)(3)),
		))
		if err != nil {
			t.Errorf("expect no error, got: %v", err)
		}
		if result != 9 {
			t.Errorf("expect 9, got: %v", result)
		}
	})
	t.Run("JoinStyleAbortOnError", func(t *testing.T) {
		errorMultiply = fmt.Errorf("some-error")
		errorAdd = nil

		result, err := Unpack(Join(
			Raise(Wrap(addThree))(Wrap(multiplyTwo)(3)),
		))
		t.Logf("got error: %v", err)
		if !errors.Is(err, errorMultiply) {
			t.Errorf("expect multiply error, got: %v", err)
		}
		if result != 0 {
			t.Errorf("expect 0, got: %v", result)
		}
	})
}

func TestMonadLongerChain(t *testing.T) {

	result, err := Unpack(Bind(Bind(Bind(
		Return(3),
		Wrap(func(in int) (int, error) { return in * 2, nil })),
		Wrap(func(in int) (int, error) { return in + 3, nil })),
		Ensure(strconv.Itoa),
	))
	if err != nil {
		t.Errorf("expect no error, got: %v", err)
	}
	if result != "9" {
		t.Errorf("expect 9, got: %v", result)
	}
}

func TestMonadWithDefault(t *testing.T) {
	t.Run("GotDefaultFromValue", func(t *testing.T) {
		result, err := Unpack(WithDefault(Bind(
			Return(3),
			Wrap(func(int) (int, error) { return 0, fmt.Errorf("some-error") })),
			FromValue(-1),
		))
		if err != nil {
			t.Errorf("expect no error, got: %v", err)
		}
		if result != -1 {
			t.Errorf("expect 0, got: %v", result)
		}
	})
	t.Run("GotDefaultFromGenerator", func(t *testing.T) {
		result, err := Unpack(WithDefault(Bind(
			Return(3),
			Wrap(func(int) (int, error) { return 0, fmt.Errorf("some-error") })),
			Build(func() (int, error) { return -1, nil }),
		))
		if err != nil {
			t.Errorf("expect no error, got: %v", err)
		}
		if result != -1 {
			t.Errorf("expect 0, got: %v", result)
		}
	})
}
