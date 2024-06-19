package cli

import "flag"

type FlagInfo = flag.Flag

type ErrorHandling = flag.ErrorHandling

const (
	ContinueOnError ErrorHandling = flag.ContinueOnError
	ExitOnError     ErrorHandling = flag.ExitOnError
	PanicOnError    ErrorHandling = flag.PanicOnError
)
