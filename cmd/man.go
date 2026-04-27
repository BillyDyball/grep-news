package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate man page",
	Long:  "Generates a man page from the CLI definition and writes it to stdout.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doc.GenMan(rootCmd, nil, os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(manCmd)
}
