package main

import (
	"fmt"
	"os"

	"github.com/cmgsj/goserve/pkg/cmd"
)

func main() {
	if err := cmd.ExecuteRootCmd(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
