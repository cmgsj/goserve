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
	var buf bytes.Buffer

	for _, config := range configs {
		if config.disabled {
			continue
		}

		if config.value == nil {
			buf.WriteString(sprintfln("  %s", config.key))
		} else {
			buf.WriteString(sprintfln("  %s:\t%v", config.key, config.value))
		}
	}

	tab := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	_, err := buf.WriteTo(tab)
	if err != nil {
		return err
	}

	return tab.Flush()
}

func printfln(format string, args ...any) {
	fmt.Fprint(os.Stdout, sprintfln(format, args...))
}

func sprintfln(format string, args ...any) string {
	return fmt.Sprintf("# "+format+"\n", args...)
}
