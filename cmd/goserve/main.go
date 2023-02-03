package main

import (
	"fmt"
	"goserve/pkg/file"
	"goserve/pkg/handler"
	"goserve/pkg/middleware"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "goserve [path]",
	Short:   "Simple static file server",
	Long:    "Simple static file server.",
	Version: "0.0.1",
	Args:    cobra.MaximumNArgs(1),
	RunE:    runRootCmd,
}

func init() {
	os.Setenv("GOSERVE_VERSION", rootCmd.Version)
	rootCmd.PersistentFlags().IntP("port", "p", 1234, "port to listen on")
	rootCmd.PersistentFlags().BoolP("text", "t", true, "serve as text or download")
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return err
	}
	text, err := cmd.Flags().GetBool("text")
	if err != nil {
		return err
	}
	var rootFile, serveMode string
	var defaultHandler http.Handler
	if len(args) == 0 {
		rootFile = "."
	} else {
		rootFile = args[0]
	}
	rootFile, err = filepath.Abs(rootFile)
	if err != nil {
		return err
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
	cmd.Printf("serving [%s] as %s at http://localhost%s\n", root.Path, serveMode, addr)
	return http.ListenAndServe(addr, defaultHandler)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
