package main

import (
	"fmt"
	"goserve/cmd/root"
	"os"
)

func main() {
	if err := root.NewCmd(root.DefaultConfig()).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
