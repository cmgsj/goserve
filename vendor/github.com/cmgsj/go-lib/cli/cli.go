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

func StringFlag(name, usage string, required bool, defaults ...string) *Flag[string] {
	return DefaultFlagSet.StringFlag(name, usage, required, defaults...)
}

func BoolFlag(name, usage string, required bool, defaults ...bool) *Flag[bool] {
	return DefaultFlagSet.BoolFlag(name, usage, required, defaults...)
}

func IntFlag(name, usage string, required bool, defaults ...int) *Flag[int] {
	return DefaultFlagSet.IntFlag(name, usage, required, defaults...)
}

func Int64Flag(name, usage string, required bool, defaults ...int64) *Flag[int64] {
	return DefaultFlagSet.Int64Flag(name, usage, required, defaults...)
}

func UintFlag(name, usage string, required bool, defaults ...uint) *Flag[uint] {
	return DefaultFlagSet.UintFlag(name, usage, required, defaults...)
}

func Uint64Flag(name, usage string, required bool, defaults ...uint64) *Flag[uint64] {
	return DefaultFlagSet.Uint64Flag(name, usage, required, defaults...)
}

func Float64Flag(name, usage string, required bool, defaults ...float64) *Flag[float64] {
	return DefaultFlagSet.Float64Flag(name, usage, required, defaults...)
}

func DurationFlag(name, usage string, required bool, defaults ...time.Duration) *Flag[time.Duration] {
	return DefaultFlagSet.DurationFlag(name, usage, required, defaults...)
}
