package main

import (
	"fmt"
	"os"

	"github.com/cmgsj/goserve/pkg/cmd/goserve"
)

func main() {
	err := goserve.NewCommandGoserve().Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
