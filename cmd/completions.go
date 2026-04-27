package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionsCmd = &cobra.Command{
	Use:   "completions [bash|zsh|fish]",
	Short: "Generate shell completions",
	Long:  "Generate shell completion scripts for bash, zsh, or fish.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		default:
			return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionsCmd)
}
