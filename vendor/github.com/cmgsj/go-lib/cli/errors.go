package cli

import (
	"flag"
	"fmt"
	"os"
)

type ErrorHandling int

const (
	ContinueOnError = ErrorHandling(flag.ContinueOnError)
	ExitOnError     = ErrorHandling(flag.ExitOnError)
	PanicOnError    = ErrorHandling(flag.PanicOnError)
)

func handleError(errorHandling ErrorHandling, err error) error {
	if err == nil {
		return nil
	}

	switch errorHandling {
	case ContinueOnError:
		return err

	case ExitOnError:
		os.Exit(2)

	case PanicOnError:
		panic(err)

	default:
		panic(fmt.Sprintf("unknown cli error handling: %d", errorHandling))
	}

	return err
}
