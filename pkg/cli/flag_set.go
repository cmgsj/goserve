package cli

import (
	"errors"
	"flag"
	"os"
	"time"
)

type FlagSet struct {
	flagSet   *flag.FlagSet
	flags     []interface{ parse() error }
	envPrefix string
}

func NewFlagSet(name string, errorHandling flag.ErrorHandling) *FlagSet {
	return &FlagSet{
		flagSet: flag.NewFlagSet(name, errorHandling),
	}
}

func (f *FlagSet) EnvPrefix() string {
	return f.envPrefix
}

func (f *FlagSet) SetEnvPrefix(envPrefix string) {
	f.envPrefix = envPrefix
}

func (f *FlagSet) Usage() {
	f.flagSet.Usage()
}

func (f *FlagSet) SetUsage(usage func(*FlagSet)) {
	f.flagSet.Usage = func() { usage(f) }
}

func (f *FlagSet) PrintDefaults() {
	f.flagSet.PrintDefaults()
}

func (f *FlagSet) NFlag() int {
	return f.flagSet.NFlag()
}

func (f *FlagSet) NArg() int {
	return f.flagSet.NArg()
}

func (f *FlagSet) Arg(i int) string {
	return f.flagSet.Arg(i)
}

func (f *FlagSet) Args() []string {
	return f.flagSet.Args()
}

func (f *FlagSet) StringFlag(name, usage string, required bool, values ...string) *Flag[string] {
	return newFlag(f, name, usage, required, values...)
}

func (f *FlagSet) BoolFlag(name, usage string, required bool, values ...bool) *Flag[bool] {
	return newFlag(f, name, usage, required, values...)
}

func (f *FlagSet) IntFlag(name, usage string, required bool, values ...int) *Flag[int] {
	return newFlag(f, name, usage, required, values...)
}

func (f *FlagSet) Int64Flag(name, usage string, required bool, values ...int64) *Flag[int64] {
	return newFlag(f, name, usage, required, values...)
}

func (f *FlagSet) UintFlag(name, usage string, required bool, values ...uint) *Flag[uint] {
	return newFlag(f, name, usage, required, values...)
}

func (f *FlagSet) Uint64Flag(name, usage string, required bool, values ...uint64) *Flag[uint64] {
	return newFlag(f, name, usage, required, values...)
}

func (f *FlagSet) Float64Flag(name, usage string, required bool, values ...float64) *Flag[float64] {
	return newFlag(f, name, usage, required, values...)
}

func (f *FlagSet) DurationFlag(name, usage string, required bool, values ...time.Duration) *Flag[time.Duration] {
	return newFlag(f, name, usage, required, values...)
}

func (f *FlagSet) Parsed() bool {
	return f.flagSet.Parsed()
}

func (f *FlagSet) Parse() error {
	err := f.flagSet.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	var errs []error

	for _, flag := range f.flags {
		errs = append(errs, flag.parse())
	}

	return errors.Join(errs...)
}
