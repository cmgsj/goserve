package cli

import (
	"os"
	"time"
)

var DefaultVars = NewVars(os.Args[0], ExitOnError)

func BindEnv() bool {
	return DefaultVars.BindEnv()
}

func SetBindEnv(bindEnv bool) {
	DefaultVars.SetBindEnv(bindEnv)
}

func EnvPrefix() string {
	return DefaultVars.EnvPrefix()
}

func SetEnvPrefix(envPrefix string) {
	DefaultVars.SetEnvPrefix(envPrefix)
}

func Name() string {
	return DefaultVars.Name()
}

func Usage() {
	DefaultVars.Usage()
}

func SetUsage(usage func(*Vars)) {
	DefaultVars.SetUsage(usage)
}

func PrintDefaults() {
	DefaultVars.PrintDefaults()
}

func Arg(i int) string {
	return DefaultVars.Arg(i)
}

func Args() []string {
	return DefaultVars.Args()
}

func NArg() int {
	return DefaultVars.NArg()
}

func Parse() error {
	return DefaultVars.Parse(os.Args[1:])
}

func Parsed() bool {
	return DefaultVars.Parsed()
}

func Bool(name, usage string, opts ...VarOption[bool]) *Var[bool] {
	return DefaultVars.Bool(name, usage, opts...)
}

func Int(name, usage string, opts ...VarOption[int]) *Var[int] {
	return DefaultVars.Int(name, usage, opts...)
}

func Int64(name, usage string, opts ...VarOption[int64]) *Var[int64] {
	return DefaultVars.Int64(name, usage, opts...)
}

func Uint(name, usage string, opts ...VarOption[uint]) *Var[uint] {
	return DefaultVars.Uint(name, usage, opts...)
}

func Uint64(name, usage string, opts ...VarOption[uint64]) *Var[uint64] {
	return DefaultVars.Uint64(name, usage, opts...)
}

func Float64(name, usage string, opts ...VarOption[float64]) *Var[float64] {
	return DefaultVars.Float64(name, usage, opts...)
}

func Duration(name, usage string, opts ...VarOption[time.Duration]) *Var[time.Duration] {
	return DefaultVars.Duration(name, usage, opts...)
}

func String(name, usage string, opts ...VarOption[string]) *Var[string] {
	return DefaultVars.String(name, usage, opts...)
}
