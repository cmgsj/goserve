package root

import (
	"fmt"
	"goserve/pkg/file"
	"goserve/pkg/format"
	"goserve/pkg/handler"
	"goserve/pkg/middleware"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

type Config struct {
	Port         int
	File         string
	SkipDotFiles bool
	LogEnabled   bool
	RawEnabled   bool
}

func NewCmd() *cobra.Command {
	config := &Config{
		Port:         1234,
		File:         ".",
		SkipDotFiles: true,
		RawEnabled:   true,
		LogEnabled:   true,
	}
	rootCmd := &cobra.Command{
		Use:     "goserve <filepath>",
		Short:   "Static file server",
		Long:    "Http static file server with web UI.",
		Version: "1.0.0",
		Args:    cobra.MaximumNArgs(1),
		RunE:    makeRunFunc(config),
	}
	rootCmd.Flags().IntVarP(&config.Port, "port", "p", config.Port, "port to listen on")
	rootCmd.Flags().BoolVar(&config.SkipDotFiles, "skip-dot-files", config.SkipDotFiles, "whether to skip files that start with \".\" or not")
	rootCmd.Flags().BoolVar(&config.LogEnabled, "log", config.LogEnabled, "whether to log request info to stdout or not")
	rootCmd.Flags().BoolVar(&config.RawEnabled, "raw", config.RawEnabled, "whether to serve raw files or to download")
	return rootCmd
}

func makeRunFunc(config *Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			config.File = args[0]
		}
		start := time.Now()
		root, numfiles, totalSize, err := file.GetFileTree(config.File, config.SkipDotFiles, cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		delta := time.Since(start)
		errch := make(chan error)
		httphandler := handler.ServeFileTree(root, config.RawEnabled, cmd.Version, errch)
		if config.LogEnabled {
			httphandler = middleware.Logger(httphandler, cmd.OutOrStdout())
		}
		addr := fmt.Sprintf(":%d", config.Port)
		printInfo(cmd, config, numfiles, totalSize, delta, root.Path, addr)
		go func() {
			for err := range errch {
				cmd.PrintErrln(err)
			}
		}()
		return http.ListenAndServe(addr, httphandler)
	}
}

func printInfo(cmd *cobra.Command, config *Config, numfiles int, totalSize int64, delta time.Duration, rootpath string, addr string) {
	cmd.Println()
	cmd.Printf("Parsed %s files [%s] in %s\n", format.ThousandsSeparator(numfiles), format.FileSize(totalSize), format.TimeDuration(delta))
	cmd.Println()
	cmd.Printf("Root: %s\n", rootpath)
	cmd.Printf("SkipDotFiles: %t\n", config.SkipDotFiles)
	cmd.Printf("RawEnabled: %t\n", config.RawEnabled)
	cmd.Printf("LogEnabled: %t\n", config.LogEnabled)
	cmd.Printf("Address: http://localhost%s\n", addr)
	cmd.Println()
	cmd.Println("Ready to accept conections")
	cmd.Println()
}
