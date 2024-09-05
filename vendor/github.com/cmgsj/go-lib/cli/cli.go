package cli

import (
	"os"
	"time"
)

var DefaultFlagSet = NewFlagSet(os.Args[0], ExitOnError)

func BindEnv() bool {
	return DefaultFlagSet.BindEnv()
}

func SetBindEnv(bindEnv bool) {
	DefaultFlagSet.SetBindEnv(bindEnv)
}

func EnvPrefix() string {
	return DefaultFlagSet.EnvPrefix()
}

func SetEnvPrefix(envPrefix string) {
	DefaultFlagSet.SetEnvPrefix(envPrefix)
}

func Usage() {
	DefaultFlagSet.Usage()
}

func SetUsage(usage func(*FlagSet)) {
	DefaultFlagSet.SetUsage(usage)
}

func PrintDefaults() {
	DefaultFlagSet.PrintDefaults()
}

func Parse() error {
	return DefaultFlagSet.Parse(os.Args[1:])
}

func Parsed() bool {
	return DefaultFlagSet.Parsed()
}

func Arg(i int) string {
	return DefaultFlagSet.Arg(i)
}

func Args() []string {
	return DefaultFlagSet.Args()
}

func NArg() int {
	return DefaultFlagSet.NArg()
}

func Set(name, value string) error {
	return DefaultFlagSet.Set(name, value)
}

func Lookup(name string) *FlagInfo {
	return DefaultFlagSet.Lookup(name)
}

func NFlag() int {
	return DefaultFlagSet.NFlag()
}

func Visit(fn func(*FlagInfo)) {
	DefaultFlagSet.Visit(fn)
}

func VisitAll(fn func(*FlagInfo)) {
	DefaultFlagSet.VisitAll(fn)
}

func Bool(name, usage string, opts ...FlagOption[bool]) *Flag[bool] {
	return DefaultFlagSet.Bool(name, usage, opts...)
}

func Int(name, usage string, opts ...FlagOption[int]) *Flag[int] {
	return DefaultFlagSet.Int(name, usage, opts...)
}

func Int64(name, usage string, opts ...FlagOption[int64]) *Flag[int64] {
	return DefaultFlagSet.Int64(name, usage, opts...)
}

func Uint(name, usage string, opts ...FlagOption[uint]) *Flag[uint] {
	return DefaultFlagSet.Uint(name, usage, opts...)
}

func Uint64(name, usage string, opts ...FlagOption[uint64]) *Flag[uint64] {
	return DefaultFlagSet.Uint64(name, usage, opts...)
}

func Float64(name, usage string, opts ...FlagOption[float64]) *Flag[float64] {
	return DefaultFlagSet.Float64(name, usage, opts...)
}

func Duration(name, usage string, opts ...FlagOption[time.Duration]) *Flag[time.Duration] {
	return DefaultFlagSet.Duration(name, usage, opts...)
}

func String(name, usage string, opts ...FlagOption[string]) *Flag[string] {
	return DefaultFlagSet.String(name, usage, opts...)
}
