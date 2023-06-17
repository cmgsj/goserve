package cmd

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cmgsj/goserve/pkg/files"
	"github.com/cmgsj/goserve/pkg/middleware"
	"github.com/cmgsj/goserve/pkg/version"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func ExecuteRootCmd() error {
	return NewRootCmd().Execute()
}

func NewRootCmd() *cobra.Command {
	c := &Config{
		Port:         1234,
		RawEnabled:   true,
		LogEnabled:   true,
		SkipDotFiles: true,
		RootPath:     ".",
	}
	cmd := &cobra.Command{
		Use:     "goserve [filepath]",
		Short:   "Static file server",
		Long:    "Http static file server with web UI.",
		Version: version.Version,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := c.Complete(cmd, args)
			if err != nil {
				return err
			}
			return c.Run(cmd, args)
		},
	}
	cmd.PersistentFlags().IntVar(&c.Port, "port", c.Port, "port to listen on")
	cmd.PersistentFlags().BoolVar(&c.RawEnabled, "raw", c.RawEnabled, "whether to serve raw files or to download")
	cmd.PersistentFlags().BoolVar(&c.LogEnabled, "log", c.LogEnabled, "whether to log request info to stdout or not")
	cmd.PersistentFlags().BoolVar(&c.SkipDotFiles, "skip-dot-files", c.SkipDotFiles, `whether to skip files whose name starts with "." or not`)
	return cmd
}

type Config struct {
	Port         int
	RawEnabled   bool
	LogEnabled   bool
	SkipDotFiles bool
	RootPath     string
}

func (c *Config) Complete(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		c.RootPath = args[0]
	}
	rootPath, err := filepath.Abs(c.RootPath)
	if err != nil {
		return err
	}
	_, err = os.Stat(rootPath)
	return err
}

func (c *Config) Run(cmd *cobra.Command, args []string) error {
	fsys := afero.NewReadOnlyFs(afero.NewBasePathFs(afero.NewOsFs(), c.RootPath))
	handler := files.NewServer(fsys, cmd.ErrOrStderr(), c.SkipDotFiles, c.RawEnabled)
	if c.LogEnabled {
		handler = middleware.NewHTTPLogger(handler, cmd.OutOrStdout())
	}
	cmd.Println()
	cmd.Printf("Root: %s\n", c.RootPath)
	cmd.Printf("SkipDotFiles: %t\n", c.SkipDotFiles)
	cmd.Printf("RawEnabled: %t\n", c.RawEnabled)
	cmd.Printf("LogEnabled: %t\n", c.LogEnabled)
	cmd.Printf("Address: http://localhost:%d\n", c.Port)
	cmd.Println()
	cmd.Println("Ready to accept connections")
	cmd.Println()
	return http.ListenAndServe(":"+strconv.Itoa(c.Port), handler)
}
