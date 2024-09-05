package cli

import "time"

type FlagOption[V FlagValue] func(*Flag[V])

func Default[V FlagValue](defaultValue V) FlagOption[V] {
	return func(f *Flag[V]) { f.defaultValue = defaultValue }
}

func RequiredInt(required bool) FlagOption[int] {
	return func(f *Flag[int]) { f.required = required }
}

func RequiredInt64(required bool) FlagOption[int64] {
	return func(f *Flag[int64]) { f.required = required }
}

func RequiredUint(required bool) FlagOption[uint] {
	return func(f *Flag[uint]) { f.required = required }
}

func RequiredUint64(required bool) FlagOption[uint64] {
	return func(f *Flag[uint64]) { f.required = required }
}

func RequiredFloat64(required bool) FlagOption[float64] {
	return func(f *Flag[float64]) { f.required = required }
}

func RequiredDuration(required bool) FlagOption[time.Duration] {
	return func(f *Flag[time.Duration]) { f.required = required }
}

func RequiredString(required bool) FlagOption[string] {
	return func(f *Flag[string]) { f.required = required }
}
