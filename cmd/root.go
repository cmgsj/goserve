package cmd

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cmgsj/goserve/pkg/handler"
	"github.com/cmgsj/goserve/pkg/middleware"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var rootCmd = newRootCmd()

func Execute() error {
	return rootCmd.Execute()
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "goserve [filepath]",
		Short:   "Static file server",
		Long:    "Http static file server with web UI.",
		Version: "1.0.0",
		Args:    cobra.MaximumNArgs(1),
		RunE:    runRootCmdE,
	}
	rootCmd.PersistentFlags().IntP("port", "p", 1234, "port to listen on")
	rootCmd.PersistentFlags().Duration("cache-time", time.Minute, "expiration time of files on the in-memory cache")
	rootCmd.PersistentFlags().Bool("raw", true, "whether to serve raw files or to download")
	rootCmd.PersistentFlags().Bool("log", true, "whether to log request info to stdout or not")
	rootCmd.PersistentFlags().Bool("skip-dot-files", true, `whether to skip files whose name starts with "." or not`)
	return rootCmd
}

func runRootCmdE(cmd *cobra.Command, args []string) error {
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return err
	}
	cacheTime, err := cmd.Flags().GetDuration("cache-time")
	if err != nil {
		return err
	}
	if cacheTime.Abs().Nanoseconds() < time.Minute.Abs().Nanoseconds() {
		cacheTime = time.Minute
	}
	rawEnabled, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return err
	}
	logEnabled, err := cmd.Flags().GetBool("log")
	if err != nil {
		return err
	}
	skipDotFiles, err := cmd.Flags().GetBool("skip-dot-files")
	if err != nil {
		return err
	}
	var rootFile = "."
	if len(args) > 0 {
		rootFile = args[0]
	}
	absPath, err := filepath.Abs(rootFile)
	if err != nil {
		return err
	}
	rootFile = absPath
	if _, err = os.Stat(rootFile); err != nil {
		return err
	}
	var (
		fsys = afero.NewCacheOnReadFs(
			afero.NewBasePathFs(
				afero.NewReadOnlyFs(afero.NewOsFs()),
				rootFile,
			),
			afero.NewMemMapFs(),
			cacheTime,
		)
		errC     = make(chan error)
		fsConfig = handler.FileServerConfig{
			FS:           fsys,
			SkipDotFiles: skipDotFiles,
			RawEnabled:   rawEnabled,
			Version:      cmd.Version,
			ErrC:         errC,
		}
		httpHandler = handler.FileServer(fsConfig)
	)
	if logEnabled {
		httpHandler = middleware.Logger(httpHandler, cmd.OutOrStdout())
	}
	go func() {
		for err := range errC {
			cmd.PrintErrln(err)
		}
	}()
	cmd.Println()
	cmd.Printf("Root: %s\n", rootFile)
	cmd.Printf("CacheTime: %s\n", cacheTime)
	cmd.Printf("SkipDotFiles: %t\n", skipDotFiles)
	cmd.Printf("RawEnabled: %t\n", rawEnabled)
	cmd.Printf("LogEnabled: %t\n", logEnabled)
	cmd.Printf("Address: http://localhost:%d\n", port)
	cmd.Println()
	cmd.Println("Ready to accept connections")
	cmd.Println()
	return http.ListenAndServe(":"+strconv.Itoa(port), httpHandler)
}