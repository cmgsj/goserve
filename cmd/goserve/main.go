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
	Use:     "goserve [filepath]",
	Short:   "Simple static file server",
	Long:    "Simple static file server.",
	Version: "0.0.1",
	Args:    cobra.MaximumNArgs(1),
	RunE:    runRootCmd,
}

func init() {
	os.Setenv("GOSERVE_VERSION", rootCmd.Version)
	rootCmd.PersistentFlags().IntP("port", "p", 1234, "port to listen on")
	rootCmd.PersistentFlags().BoolP("raw", "r", true, "serve raw content or to download")
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return err
	}
	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return err
	}
	var (
		fpath       string
		serveMode   string
		httphandler http.Handler
	)
	if len(args) == 0 {
		fpath = "."
	} else {
		fpath = args[0]
	}
	fpath, err = filepath.Abs(fpath)
	if err != nil {
		return err
	}
	root, err := file.GetFileTree(cmd.ErrOrStderr(), fpath)
	if err != nil {
		return err
	}
	httphandler = middleware.Logger(cmd.OutOrStdout(), handler.ServeFileTree(root, raw))
	if raw {
		serveMode = "raw"
	} else {
		serveMode = "download"
	}
	addr := fmt.Sprintf(":%d", port)
	cmd.Printf("serving %s [%s] at http://localhost%s\n", serveMode, root.Path, addr)
	return http.ListenAndServe(addr, httphandler)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
