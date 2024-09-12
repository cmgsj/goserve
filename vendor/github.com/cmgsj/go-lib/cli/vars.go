package cli

import (
	"errors"
	"flag"
	"time"
)

type Vars struct {
	flagSet       *flag.FlagSet
	errorHandling ErrorHandling
	vars          []interface{ parse() error }
	envPrefix     *string
}

func NewVars(name string, errorHandling ErrorHandling) *Vars {
	return &Vars{
		flagSet:       flag.NewFlagSet(name, flag.ErrorHandling(errorHandling)),
		errorHandling: errorHandling,
	}
}

func (v *Vars) BindEnv() bool {
	return v.envPrefix != nil
}

func (v *Vars) SetBindEnv(bindEnv bool) {
	if !bindEnv {
		v.envPrefix = nil
	} else if v.envPrefix == nil {
		v.envPrefix = new(string)
	}
}

func (v *Vars) EnvPrefix() string {
	if v.envPrefix != nil {
		return *v.envPrefix
	}
	return ""
}

func (v *Vars) SetEnvPrefix(envPrefix string) {
	v.envPrefix = &envPrefix
}

func (v *Vars) Name() string {
	return v.flagSet.Name()
}

func (v *Vars) Usage() {
	v.flagSet.Usage()
}

func (v *Vars) SetUsage(usage func(*Vars)) {
	v.flagSet.Usage = func() { usage(v) }
}

func (v *Vars) PrintDefaults() {
	v.flagSet.PrintDefaults()
}

func (v *Vars) Arg(i int) string {
	return v.flagSet.Arg(i)
}

func (v *Vars) Args() []string {
	return v.flagSet.Args()
}

func (v *Vars) NArg() int {
	return v.flagSet.NArg()
}

func (v *Vars) Parsed() bool {
	return v.flagSet.Parsed()
}

func (v *Vars) Parse(args []string) error {
	if v.flagSet == nil && v.envPrefix == nil {
		return nil
	}

	err := v.flagSet.Parse(args)
	if err != nil {
		return handleError(v.errorHandling, err)
	}

	var errs []error

	for _, v := range v.vars {
		errs = append(errs, v.parse())
	}

	return handleError(v.errorHandling, errors.Join(errs...))
}

func (v *Vars) Bool(name, usage string, opts ...VarOption[bool]) *Var[bool] {
	return newVar(v, name, usage, opts...)
}

func (v *Vars) Int(name, usage string, opts ...VarOption[int]) *Var[int] {
	return newVar(v, name, usage, opts...)
}

func (v *Vars) Int64(name, usage string, opts ...VarOption[int64]) *Var[int64] {
	return newVar(v, name, usage, opts...)
}

func (v *Vars) Uint(name, usage string, opts ...VarOption[uint]) *Var[uint] {
	return newVar(v, name, usage, opts...)
}

func (v *Vars) Uint64(name, usage string, opts ...VarOption[uint64]) *Var[uint64] {
	return newVar(v, name, usage, opts...)
}

func (v *Vars) Float64(name, usage string, opts ...VarOption[float64]) *Var[float64] {
	return newVar(v, name, usage, opts...)
}

func (v *Vars) Duration(name, usage string, opts ...VarOption[time.Duration]) *Var[time.Duration] {
	return newVar(v, name, usage, opts...)
}

func (v *Vars) String(name, usage string, opts ...VarOption[string]) *Var[string] {
	return newVar(v, name, usage, opts...)
}
