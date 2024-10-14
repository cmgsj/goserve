package goserve

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type loggerOptions struct {
	format string
	level  string
}

func initLogger(o loggerOptions) error {
	var level slog.Level

	err := level.UnmarshalText([]byte(o.level))
	if err != nil {
		return err
	}

	out := os.Stdout

	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch strings.ToLower(o.format) {
	case "json":
		handler = slog.NewJSONHandler(out, opts)

	case "text":
		handler = slog.NewTextHandler(out, opts)

	default:
		return fmt.Errorf("invalid log format %q", o.format)
	}

	slog.SetDefault(slog.New(handler))

	return nil
}
