package goserve

import (
	"bytes"
	"fmt"
	"os"
	"text/tabwriter"
)

type config struct {
	key      string
	value    any
	disabled bool
}

func printConfigs(configs []config) error {
	var lines []string

	for _, config := range configs {
		if config.disabled {
			continue
		}

		if config.value == nil {
			lines = append(lines, sprintfln("  %s", config.key))
		} else {
			lines = append(lines, sprintfln("  %s:\t%v", config.key, config.value))
		}
	}

	return tabwrite(lines)
}

func tabwrite(lines []string) error {
	var buf bytes.Buffer

	for _, line := range lines {
		buf.WriteString(line)
	}

	tab := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	_, err := buf.WriteTo(tab)
	if err != nil {
		return err
	}

	return tab.Flush()
}

const printPrefix = "# "

func println(args ...any) {
	if !silent.Value() {
		fmt.Println(printPrefix + fmt.Sprint(args...))
	}
}

func printfln(format string, args ...any) {
	if !silent.Value() {
		fmt.Printf(printPrefix+format+"\n", args...)
	}
}

func sprintfln(format string, args ...any) string {
	if !silent.Value() {
		return fmt.Sprintf(printPrefix+format+"\n", args...)
	}
	return ""
}
