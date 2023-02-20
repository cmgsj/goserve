package root

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

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
	var (
		cfg = &config{
			Port:         1234,
			File:         ".",
			SkipDotFiles: true,
			RawEnabled:   true,
			LogEnabled:   true,
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
	rootCmd.PersistentFlags().IntVarP(&cfg.Port, "port", "p", cfg.Port, "port to listen on")
	rootCmd.PersistentFlags().BoolVar(&cfg.SkipDotFiles, "skip-dot-files", cfg.SkipDotFiles, `whether to skip files whose name starts with "." or not`)
	rootCmd.PersistentFlags().BoolVar(&cfg.LogEnabled, "log", cfg.LogEnabled, "whether to log request info to stdout or not")
	rootCmd.PersistentFlags().BoolVar(&cfg.RawEnabled, "raw", cfg.RawEnabled, "whether to serve raw files or to download")
	return rootCmd
}

func makeRunFunc(cfg *config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			cfg.File = args[0]
		}
		addr := ":" + strconv.Itoa(cfg.Port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		defer lis.Close()
		root, stats, err := file.GetFileTree(cfg.File, cfg.SkipDotFiles, cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		var (
			errCh       = make(chan error)
			httpHandler = handler.ServeFileTree(root, cfg.RawEnabled, cmd.Version, errCh)
		)
		if cfg.LogEnabled {
			httpHandler = middleware.Logger(httpHandler, cmd.OutOrStdout())
		}
		printInfo(cmd, cfg, stats, root.Path, addr)
		go func() {
			for err := range errCh {
				cmd.PrintErrln(err)
			}
		}()
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		go func() {
			<-sigs
			cmd.Println()
			cmd.Println("Shutting down...")
			os.Exit(0)
		}()
		return http.Serve(lis, httpHandler)
	}
}

func printInfo(cmd *cobra.Command, cfg *config, stats *file.TreeStats, rootPath string, addr string) {
	cmd.Println()
	cmd.Printf("Parsed %s files [%s] in %s\n",
		format.ThousandsSeparator(stats.NumFiles), format.FileSize(stats.TotalSize), format.Duration(stats.TimeDelta))
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
