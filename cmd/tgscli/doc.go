package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	docDirFlag string
)

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.Flags().StringVarP(&docDirFlag, "doc-path", "", "./docs", "Path directory where you want generate doc files")
}

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "gen docs",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := os.MkdirAll(docDirFlag, 0644)
		if err != nil {
			errorExitf("Generate docs failed: %s", err)
		}

		err = doc.GenMarkdownTree(rootCmd, docDirFlag)
		if err != nil {
			errorExitf("Generate docs failed: %s", err)
		}
	},
}
