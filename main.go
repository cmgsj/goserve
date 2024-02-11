package main

import (
	"fmt"
	"os"

	"github.com/cmgsj/goserve/pkg/cmd"
)

func main() {
	err := cmd.NewRootCmd().Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
