package root

import (
	"fmt"
	"net"
	"net/http"

	"github.com/cmgsj/goserve/pkg/file"
	"github.com/cmgsj/goserve/pkg/format"
	"github.com/cmgsj/goserve/pkg/handler"
	"github.com/cmgsj/goserve/pkg/middleware"
	"github.com/spf13/cobra"
)

var rootCmd = newRootCmd()

func Execute() error {
	return rootCmd.Execute()
}

type config struct {
	Port         int
	File         string
	SkipDotFiles bool
	LogEnabled   bool
	RawEnabled   bool
}

func newRootCmd() *cobra.Command {
	cfg := &config{
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
		RunE:    makeRunFunc(cfg),
	}
	rootCmd.PersistentFlags().IntVarP(&cfg.Port, "port", "p", cfg.Port, "port to listen on")
	rootCmd.PersistentFlags().BoolVar(&cfg.SkipDotFiles, "skip-dot-files", cfg.SkipDotFiles, "whether to skip files that start with \".\" or not")
	rootCmd.PersistentFlags().BoolVar(&cfg.LogEnabled, "log", cfg.LogEnabled, "whether to log request info to stdout or not")
	rootCmd.PersistentFlags().BoolVar(&cfg.RawEnabled, "raw", cfg.RawEnabled, "whether to serve raw files or to download")
	return rootCmd
}

func makeRunFunc(cfg *config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			cfg.File = args[0]
		}
		addr := fmt.Sprintf(":%d", cfg.Port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		defer lis.Close()
		root, info, err := file.GetFileTree(cfg.File, cfg.SkipDotFiles, cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		errCh := make(chan error)
		httpHandler := handler.ServeFileTree(root, cfg.RawEnabled, cmd.Version, errCh)
		if cfg.LogEnabled {
			httpHandler = middleware.Logger(httpHandler, cmd.OutOrStdout())
		}
		printInfo(cmd, cfg, info, root.Path, addr)
		go func() {
			for err := range errCh {
				cmd.PrintErrln(err)
			}
		}()
		return http.Serve(lis, httpHandler)
	}
}

func printInfo(cmd *cobra.Command, cfg *config, info *file.TreeInfo, rootPath string, addr string) {
	cmd.Println()
	cmd.Printf("Parsed %s files [%s] in %s\n",
		format.ThousandsSeparator(info.NumFiles), format.FileSize(info.TotalSize), format.Duration(info.TimeDelta))
	cmd.Println()
	cmd.Printf("Root: %s\n", rootPath)
	cmd.Printf("SkipDotFiles: %t\n", cfg.SkipDotFiles)
	cmd.Printf("RawEnabled: %t\n", cfg.RawEnabled)
	cmd.Printf("LogEnabled: %t\n", cfg.LogEnabled)
	cmd.Printf("Address: http://localhost%s\n", addr)
	cmd.Println()
	cmd.Println("Ready to accept connections")
	cmd.Println()
}
