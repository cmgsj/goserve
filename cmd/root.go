package cmd

import (
	"errors"
	"fmt"
	"goserve/pkg/handler"
	"goserve/pkg/middleware"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var (
	port    int
	text    bool
	output  string
	rootCmd = &cobra.Command{
		Use:     "goserve [path]",
		Short:   "Simple static file server",
		Long:    "Simple static file server.",
		Version: "0.1.0",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) > 0 {
				root = path.Clean(args[0])
			}
			if output != "" {
				f, err := os.Create(output)
				if err != nil {
					return err
				}
				defer f.Close()
				cmd.SetOut(io.MultiWriter(os.Stdout, f))
			}
			fstat, err := os.Stat(root)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("path %s does not exist", root)
				}
				return fmt.Errorf("reading path %s failed: %v", root, err)
			}
			defaultHandler := handler.ServeDir(root, text)
			fileType := "dir"
			if !fstat.IsDir() {
				defaultHandler = handler.ServeFile(root, fstat.Size(), text)
				fileType = "file"
			}
			defaultHandler = middleware.Logger(cmd.OutOrStdout(), defaultHandler)
			serveMode := "text"
			if !text {
				serveMode = "download"
			}
			addr := fmt.Sprintf(":%d", port)
			cmd.Printf("serving %s[%s] as %s at http://localhost%s\n", fileType, root, serveMode, addr)
			return http.ListenAndServe(addr, defaultHandler)
		},
	}
)

func Execute() {
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 1234, "port to listen on")
	rootCmd.PersistentFlags().BoolVarP(&text, "text", "t", true, "serve as text or download")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output file path")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
