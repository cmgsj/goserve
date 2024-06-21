package cli

import (
	"errors"
	"flag"
	"time"
)

type FlagSet struct {
	flagSet   *flag.FlagSet
	flags     []interface{ parse() error }
	envPrefix *string
}

func NewFlagSet(name string, errorHandling ErrorHandling) *FlagSet {
	return &FlagSet{
		flagSet: flag.NewFlagSet(name, errorHandling),
	}
}

func (f *FlagSet) BindEnv() bool {
	return f.envPrefix != nil
}

func (f *FlagSet) SetBindEnv(bindEnv bool) {
	if bindEnv {
		f.envPrefix = new(string)
	} else {
		f.envPrefix = nil
	}
}

func (f *FlagSet) EnvPrefix() string {
	if f.envPrefix != nil {
		return *f.envPrefix
	}
	return ""
}

func (f *FlagSet) SetEnvPrefix(envPrefix string) {
	f.envPrefix = &envPrefix
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

func (f *FlagSet) Parse(args []string) error {
	return f.parse(args)
}

func (f *FlagSet) Parsed() bool {
	return f.flagSet.Parsed()
}

func (f *FlagSet) Arg(i int) string {
	return f.flagSet.Arg(i)
}

func (f *FlagSet) Args() []string {
	return f.flagSet.Args()
}

func (f *FlagSet) NArg() int {
	return f.flagSet.NArg()
}

func (f *FlagSet) Set(name, value string) error {
	return f.flagSet.Set(name, value)
}

func (f *FlagSet) Lookup(name string) *FlagInfo {
	return f.flagSet.Lookup(name)
}

func (f *FlagSet) NFlag() int {
	return f.flagSet.NFlag()
}

func (f *FlagSet) Visit(fn func(*FlagInfo)) {
	f.flagSet.Visit(fn)
}

func (f *FlagSet) VisitAll(fn func(*FlagInfo)) {
	f.flagSet.VisitAll(fn)
}

func (f *FlagSet) StringFlag(name, usage string, required bool, defaults ...string) *Flag[string] {
	return newFlag(f, name, usage, required, defaults...)
}

func (f *FlagSet) BoolFlag(name, usage string, required bool, defaults ...bool) *Flag[bool] {
	return newFlag(f, name, usage, required, defaults...)
}

func (f *FlagSet) IntFlag(name, usage string, required bool, defaults ...int) *Flag[int] {
	return newFlag(f, name, usage, required, defaults...)
}

func (f *FlagSet) Int64Flag(name, usage string, required bool, defaults ...int64) *Flag[int64] {
	return newFlag(f, name, usage, required, defaults...)
}

func (f *FlagSet) UintFlag(name, usage string, required bool, defaults ...uint) *Flag[uint] {
	return newFlag(f, name, usage, required, defaults...)
}

func (f *FlagSet) Uint64Flag(name, usage string, required bool, defaults ...uint64) *Flag[uint64] {
	return newFlag(f, name, usage, required, defaults...)
}

func (f *FlagSet) Float64Flag(name, usage string, required bool, defaults ...float64) *Flag[float64] {
	return newFlag(f, name, usage, required, defaults...)
}

func (f *FlagSet) DurationFlag(name, usage string, required bool, defaults ...time.Duration) *Flag[time.Duration] {
	return newFlag(f, name, usage, required, defaults...)
}

func (f *FlagSet) parse(args []string) error {
	err := f.flagSet.Parse(args)
	if err != nil {
		return err
	}

	var errs []error

	for _, flag := range f.flags {
		errs = append(errs, flag.parse())
	}

	return errors.Join(errs...)
}
