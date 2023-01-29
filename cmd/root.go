package cmd

import (
	"fmt"
	"goserve/pkg/file"
	"goserve/pkg/handler"
	"goserve/pkg/middleware"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var (
	port    int
	text    bool
	rootCmd = &cobra.Command{
		Use:     "goserve [path]",
		Short:   "Simple static file server",
		Long:    "Simple static file server.",
		Version: "0.0.1",
		Args:    cobra.MaximumNArgs(1),
		RunE:    runRootCmd,
	}
)

func Execute() {
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 1234, "port to listen on")
	rootCmd.PersistentFlags().BoolVarP(&text, "text", "t", true, "serve as text or download")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	var rootFile, serveMode string
	var defaultHandler http.Handler
	if len(args) == 0 {
		rootFile = "."
	} else {
		rootFile = path.Clean(args[0])
	}
	if strings.Contains(rootFile, "..") {
		return fmt.Errorf("invalid path: %s (must not contain '..')", rootFile)
	}
	root, err := file.GetFSRoot(rootFile)
	if err != nil {
		return err
	}
	defaultHandler = middleware.Logger(cmd.OutOrStdout(), handler.ServeRoot(root, text))
	if text {
		serveMode = "text"
	} else {
		serveMode = "download"
	}
	addr := fmt.Sprintf(":%d", port)
	cmd.Printf("serving [%s] as %s at http://localhost%s\n", rootFile, serveMode, addr)
	return http.ListenAndServe(addr, defaultHandler)
}
