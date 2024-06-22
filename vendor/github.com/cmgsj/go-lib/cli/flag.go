package cli

import (
	"cmp"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type FlagType interface {
	string | bool | int | int64 | uint | uint64 | float64 | time.Duration
}

type Flag[T FlagType] struct {
	flagSet      *FlagSet
	name         string
	usage        string
	required     bool
	defaultValue T
	value        T
}

func newFlag[T FlagType](flagSet *FlagSet, name, usage string, required bool, defaults ...T) *Flag[T] {
	f := &Flag[T]{
		flagSet:      flagSet,
		name:         name,
		usage:        usage,
		required:     required,
		defaultValue: cmp.Or(defaults...),
	}

	f.flagSet.flags = append(f.flagSet.flags, f)

	switch any(f.value).(type) {
	case string:
		f.flagSet.flagSet.StringVar(any(&f.value).(*string), f.name, any(f.value).(string), f.usage)

	case bool:
		f.flagSet.flagSet.BoolVar(any(&f.value).(*bool), f.name, any(f.value).(bool), f.usage)

	case int:
		f.flagSet.flagSet.IntVar(any(&f.value).(*int), f.name, any(f.value).(int), f.usage)

	case int64:
		f.flagSet.flagSet.Int64Var(any(&f.value).(*int64), f.name, any(f.value).(int64), f.usage)

	case uint:
		f.flagSet.flagSet.UintVar(any(&f.value).(*uint), f.name, any(f.value).(uint), f.usage)

	case uint64:
		f.flagSet.flagSet.Uint64Var(any(&f.value).(*uint64), f.name, any(f.value).(uint64), f.usage)

	case float64:
		f.flagSet.flagSet.Float64Var(any(&f.value).(*float64), f.name, any(f.value).(float64), f.usage)

	case time.Duration:
		f.flagSet.flagSet.DurationVar(any(&f.value).(*time.Duration), f.name, any(f.value).(time.Duration), f.usage)

	default:
		panic(fmt.Sprintf("unable to add flag of type %T", f.value))
	}

	return f
}

func (f *Flag[T]) Name() string {
	return f.name
}

func (f *Flag[T]) Usage() string {
	return f.usage
}

func (f *Flag[T]) Required() bool {
	return f.required
}

func (f *Flag[T]) Value() T {
	return f.value
}

func (f *Flag[T]) SetValue(value T) {
	f.value = value
}

func (f *Flag[T]) parse() error {
	var zero T
	var boundEnv bool

	if f.value == zero && f.flagSet.envPrefix != nil {
		key := f.name

		if *f.flagSet.envPrefix != "" {
			key = *f.flagSet.envPrefix + "_" + key
		}

		key = strings.ToUpper(toSnakeCase(key))

		value, ok := os.LookupEnv(key)
		if ok {
			var v any
			var err error

			switch any(f.value).(type) {
			case string:
				v = value

			case bool:
				v, err = strconv.ParseBool(value)

			case int:
				var i int64
				i, err = strconv.ParseInt(value, 10, 0)
				v = int(i)

			case int64:
				v, err = strconv.ParseInt(value, 10, 64)

			case uint:
				var u uint64
				u, err = strconv.ParseUint(value, 10, 0)
				v = uint(u)

			case uint64:
				v, err = strconv.ParseUint(value, 10, 64)

			case float64:
				v, err = strconv.ParseFloat(value, 64)

			case time.Duration:
				v, err = time.ParseDuration(value)

			default:
				panic(fmt.Sprintf("unable to parse flag of type %T", f.value))
			}

			if err != nil {
				return fmt.Errorf(
					"failed to parse flag %s of type %T from environment variable %s=%q: %v",
					f.name, f.value, key, value, err)
			}

			f.value = v.(T)

			boundEnv = true
		}
	}

	if f.value == zero && !boundEnv {
		f.value = f.defaultValue
	}

	if f.value == zero && f.required {
		return fmt.Errorf("missing required flag %s", f.name)
	}

	return nil
}

var (
	camelCaseLowerToUpper = regexp.MustCompile("([a-z0-9])([A-Z])")
	camelCaseUpperToLower = regexp.MustCompile("([A-Z])([A-Z][a-z0-9])")
)

func toSnakeCase(s string) string {
	s = camelCaseLowerToUpper.ReplaceAllString(s, "${1}_${2}")
	s = camelCaseUpperToLower.ReplaceAllString(s, "${1}_${2}")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}
