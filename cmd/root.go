package cmd

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cmgsj/goserve/pkg/files"
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
		Version: "1.0.3",
		Args:    cobra.MaximumNArgs(1),
		RunE:    runRootCmdE,
	}
	rootCmd.PersistentFlags().Int("port", 1234, "port to listen on")
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
	rootPath := "."
	if len(args) > 0 {
		rootPath = args[0]
	}
	rootPath, err = filepath.Abs(rootPath)
	if err != nil {
		return err
	}
	if _, err = os.Stat(rootPath); err != nil {
		return err
	}
	var (
		errC    = make(chan error, 1)
		handler = files.NewServer(files.ServerConfig{
			Fs:           afero.NewBasePathFs(afero.NewReadOnlyFs(afero.NewOsFs()), rootPath),
			SkipDotFiles: skipDotFiles,
			RawEnabled:   rawEnabled,
			Version:      cmd.Version,
			ErrC:         errC,
		})
	)
	if logEnabled {
		handler = middleware.NewHTTPLogger(handler, cmd.OutOrStdout())
	}
	go func() {
		for err := range errC {
			cmd.PrintErrln(err)
		}
	}()
	cmd.Println()
	cmd.Printf("Root: %s\n", rootPath)
	cmd.Printf("SkipDotFiles: %t\n", skipDotFiles)
	cmd.Printf("RawEnabled: %t\n", rawEnabled)
	cmd.Printf("LogEnabled: %t\n", logEnabled)
	cmd.Printf("Address: http://localhost:%d\n", port)
	cmd.Println()
	cmd.Println("Ready to accept connections")
	cmd.Println()
	return http.ListenAndServe(":"+strconv.Itoa(port), handler)
}
