package cli

import (
	"flag"
	"os"
	"time"
)

var defaultFlagSet = NewFlagSet(os.Args[0], flag.ExitOnError)

func Default() *FlagSet {
	return defaultFlagSet
}

func EnvPrefix() string {
	return defaultFlagSet.EnvPrefix()
}

func SetEnvPrefix(envPrefix string) {
	defaultFlagSet.SetEnvPrefix(envPrefix)
}

func Usage() {
	defaultFlagSet.Usage()
}

func SetUsage(usage func(*FlagSet)) {
	defaultFlagSet.SetUsage(usage)
}

func PrintDefaults() {
	defaultFlagSet.PrintDefaults()
}

func NFlag() int {
	return defaultFlagSet.NFlag()
}

func NArg() int {
	return defaultFlagSet.NArg()
}

func Arg(i int) string {
	return defaultFlagSet.Arg(i)
}

func Args() []string {
	return defaultFlagSet.Args()
}

func StringFlag(name, usage string, required bool, values ...string) *Flag[string] {
	return defaultFlagSet.StringFlag(name, usage, required, values...)
}

func BoolFlag(name, usage string, required bool, values ...bool) *Flag[bool] {
	return defaultFlagSet.BoolFlag(name, usage, required, values...)
}

func IntFlag(name, usage string, required bool, values ...int) *Flag[int] {
	return defaultFlagSet.IntFlag(name, usage, required, values...)
}

func Int64Flag(name, usage string, required bool, values ...int64) *Flag[int64] {
	return defaultFlagSet.Int64Flag(name, usage, required, values...)
}

func UintFlag(name, usage string, required bool, values ...uint) *Flag[uint] {
	return defaultFlagSet.UintFlag(name, usage, required, values...)
}

func Uint64Flag(name, usage string, required bool, values ...uint64) *Flag[uint64] {
	return defaultFlagSet.Uint64Flag(name, usage, required, values...)
}

func Float64Flag(name, usage string, required bool, values ...float64) *Flag[float64] {
	return defaultFlagSet.Float64Flag(name, usage, required, values...)
}

func DurationFlag(name, usage string, required bool, values ...time.Duration) *Flag[time.Duration] {
	return defaultFlagSet.DurationFlag(name, usage, required, values...)
}

func Parsed() bool {
	return defaultFlagSet.Parsed()
}

func Parse() error {
	return defaultFlagSet.Parse()
}
