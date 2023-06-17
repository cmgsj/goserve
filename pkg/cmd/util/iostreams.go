package util

import (
	"io"

	"github.com/spf13/cobra"
)

type IOStreams interface {
	In() io.Reader
	Out() io.Writer
	Err() io.Writer
}

func NewIOStreams(in io.Reader, out, err io.Writer) IOStreams {
	return &ioStreams{in: in, out: out, err: err}
}

func NewIOStreamsFromCmd(cmd *cobra.Command) IOStreams {
	return &ioStreams{in: cmd.InOrStdin(), out: cmd.OutOrStdout(), err: cmd.ErrOrStderr()}
}

type ioStreams struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

func (s *ioStreams) In() io.Reader { return s.in }

func (s *ioStreams) Out() io.Writer { return s.out }

func (s *ioStreams) Err() io.Writer { return s.err }
