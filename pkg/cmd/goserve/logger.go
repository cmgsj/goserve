package goserve

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

func initLogger() error {
	var level slog.Level

	err := level.UnmarshalText([]byte(logLevel.Value()))
	if err != nil {
		return err
	}

	var out io.Writer

	switch strings.ToLower(logOutput.Value()) {
	case "stdout":
		out = os.Stdout
	case "stderr":
		out = os.Stderr
	default:
		out, err = os.Create(logOutput.Value())
		if err != nil {
			return err
		}
	}

	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch strings.ToLower(logFormat.Value()) {
	case "json":
		handler = slog.NewJSONHandler(out, opts)
	case "text":
		handler = slog.NewTextHandler(out, opts)
	default:
		return fmt.Errorf("invalid log format %q", logFormat.Value())
	}

	slog.SetDefault(slog.New(handler))

	return nil
}
