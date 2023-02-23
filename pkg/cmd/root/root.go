package root

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cmgsj/goserve/pkg/handler"
	"github.com/cmgsj/goserve/pkg/middleware"
	"github.com/spf13/cobra"
)

var rootCmd = newRootCmd()

func Execute() error {
	return rootCmd.Execute()
}

type config struct {
	RootFile     string
	SkipDotFiles bool
	RawEnabled   bool
	LogEnabled   bool
	Port         int
}

func newRootCmd() *cobra.Command {
	var (
		cfg = &config{
			RootFile:     ".",
			SkipDotFiles: true,
			RawEnabled:   true,
			LogEnabled:   true,
			Port:         1234,
		}
		rootCmd = &cobra.Command{
			Use:     "goserve <filepath>",
			Short:   "Static file server",
			Long:    "Http static file server with web UI.",
			Version: "1.0.0",
			Args:    cobra.MaximumNArgs(1),
			RunE:    makeRunFunc(cfg),
		}
	)
	rootCmd.PersistentFlags().BoolVar(&cfg.SkipDotFiles, "skip-dot-files", cfg.SkipDotFiles, `whether to skip files whose name starts with "." or not`)
	rootCmd.PersistentFlags().BoolVar(&cfg.RawEnabled, "raw", cfg.RawEnabled, "whether to serve raw files or to download")
	rootCmd.PersistentFlags().BoolVar(&cfg.LogEnabled, "log", cfg.LogEnabled, "whether to log request info to stdout or not")
	rootCmd.PersistentFlags().IntVarP(&cfg.Port, "port", "p", cfg.Port, "port to listen on")
	return rootCmd
}

func makeRunFunc(cfg *config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			cfg.RootFile = args[0]
		}
		absPath, err := filepath.Abs(cfg.RootFile)
		if err != nil {
			return err
		}
		cfg.RootFile = absPath
		if _, err = os.Stat(cfg.RootFile); err != nil {
			return err
		}
		var (
			errCh       = make(chan error)
			httpHandler = handler.ServeFile(cfg.RootFile, cfg.SkipDotFiles, cfg.RawEnabled, cmd.Version, errCh)
		)
		if cfg.LogEnabled {
			httpHandler = middleware.Logger(httpHandler, cmd.OutOrStdout())
		}
		go func() {
			for err := range errCh {
				cmd.PrintErrln(err)
			}
		}()
		addr := ":" + strconv.Itoa(cfg.Port)
		printInfo(cmd, cfg, addr)
		return http.ListenAndServe(addr, httpHandler)
	}
}

func printInfo(cmd *cobra.Command, cfg *config, addr string) {
	cmd.Println()
	cmd.Printf("Root: %s\n", cfg.RootFile)
	cmd.Printf("SkipDotFiles: %t\n", cfg.SkipDotFiles)
	cmd.Printf("RawEnabled: %t\n", cfg.RawEnabled)
	cmd.Printf("LogEnabled: %t\n", cfg.LogEnabled)
	cmd.Printf("Address: http://localhost%s\n", addr)
	cmd.Println()
	cmd.Println("Ready to accept connections")
	cmd.Println()
}
