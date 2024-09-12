package cli

import "time"

type VarOption[V Value] func(*Var[V])

func Default[V Value](defaultValue V) VarOption[V] {
	return func(f *Var[V]) { f.defaultValue = defaultValue }
}

func RequiredBool(required bool) VarOption[bool] {
	return func(f *Var[bool]) { f.required = required }
}

func RequiredInt(required bool) VarOption[int] {
	return func(f *Var[int]) { f.required = required }
}

func RequiredInt64(required bool) VarOption[int64] {
	return func(f *Var[int64]) { f.required = required }
}

func RequiredUint(required bool) VarOption[uint] {
	return func(f *Var[uint]) { f.required = required }
}

func RequiredUint64(required bool) VarOption[uint64] {
	return func(f *Var[uint64]) { f.required = required }
}

func RequiredFloat64(required bool) VarOption[float64] {
	return func(f *Var[float64]) { f.required = required }
}

func RequiredDuration(required bool) VarOption[time.Duration] {
	return func(f *Var[time.Duration]) { f.required = required }
}

func RequiredString(required bool) VarOption[string] {
	return func(f *Var[string]) { f.required = required }
}
