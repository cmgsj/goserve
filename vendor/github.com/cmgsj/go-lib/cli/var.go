package cli

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Value interface {
	string | bool | int | int64 | uint | uint64 | float64 | time.Duration
}

type Var[V Value] struct {
	vars         *Vars
	name         string
	usage        string
	required     bool
	defaultValue V
	value        V
}

func newVar[V Value](vars *Vars, name, usage string, opts ...VarOption[V]) *Var[V] {
	v := &Var[V]{
		vars:  vars,
		name:  name,
		usage: usage,
	}

	for _, opt := range opts {
		opt(v)
	}

	if v.vars.flagSet != nil {
		switch any(v.value).(type) {
		case bool:
			v.vars.flagSet.BoolVar(any(&v.value).(*bool), v.name, any(v.defaultValue).(bool), v.usage)

		case int:
			v.vars.flagSet.IntVar(any(&v.value).(*int), v.name, any(v.defaultValue).(int), v.usage)

		case int64:
			v.vars.flagSet.Int64Var(any(&v.value).(*int64), v.name, any(v.defaultValue).(int64), v.usage)

		case uint:
			v.vars.flagSet.UintVar(any(&v.value).(*uint), v.name, any(v.defaultValue).(uint), v.usage)

		case uint64:
			v.vars.flagSet.Uint64Var(any(&v.value).(*uint64), v.name, any(v.defaultValue).(uint64), v.usage)

		case float64:
			v.vars.flagSet.Float64Var(any(&v.value).(*float64), v.name, any(v.defaultValue).(float64), v.usage)

		case time.Duration:
			v.vars.flagSet.DurationVar(any(&v.value).(*time.Duration), v.name, any(v.defaultValue).(time.Duration), v.usage)

		case string:
			v.vars.flagSet.StringVar(any(&v.value).(*string), v.name, any(v.defaultValue).(string), v.usage)

		default:
			panic(fmt.Sprintf("unable to add variable of type %T", v.value))
		}
	}

	v.vars.vars = append(v.vars.vars, v)

	return v
}

func (v *Var[V]) Name() string {
	return v.name
}

func (v *Var[V]) Usage() string {
	return v.usage
}

func (v *Var[V]) Required() bool {
	return v.required
}

func (v *Var[V]) Value() V {
	return v.value
}

func (v *Var[V]) SetValue(value V) {
	v.value = value
}

func (v *Var[V]) parse() error {
	var flagIsSet bool

	if v.vars.flagSet != nil {
		v.vars.flagSet.Visit(func(f *flag.Flag) {
			if f.Name == v.name {
				flagIsSet = true
			}
		})
	}

	if flagIsSet {
		return nil
	}

	if v.vars.envPrefix != nil {
		key := v.name

		if *v.vars.envPrefix != "" {
			key = *v.vars.envPrefix + "_" + key
		}

		key = strings.ToUpper(toSnakeCase(key))

		value, ok := os.LookupEnv(key)
		if ok {
			var val any
			var err error

			switch any(v.value).(type) {
			case bool:
				val, err = strconv.ParseBool(value)

			case int:
				var i int64
				i, err = strconv.ParseInt(value, 10, 0)
				val = int(i)

			case int64:
				val, err = strconv.ParseInt(value, 10, 64)

			case uint:
				var u uint64
				u, err = strconv.ParseUint(value, 10, 0)
				val = uint(u)

			case uint64:
				val, err = strconv.ParseUint(value, 10, 64)

			case float64:
				val, err = strconv.ParseFloat(value, 64)

			case time.Duration:
				val, err = time.ParseDuration(value)

			case string:
				val = value

			default:
				panic(fmt.Sprintf("unable to parse variable of type %T", v.value))
			}

			if err != nil {
				return fmt.Errorf("failed to parse variable %s of type %T from environment variable %s=%q: %v", v.name, v.value, key, value, err)
			}

			v.value = val.(V)

			return nil
		}
	}

	if v.required {
		return fmt.Errorf("missing required variable %s", v.name)
	}

	return nil
}

var (
	camelLowerToUpper = regexp.MustCompile("([a-z0-9])([A-Z])")
	camelUpperToLower = regexp.MustCompile("([A-Z])([A-Z][a-z0-9])")
)

func toSnakeCase(s string) string {
	s = camelLowerToUpper.ReplaceAllString(s, "${1}_${2}")
	s = camelUpperToLower.ReplaceAllString(s, "${1}_${2}")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}
