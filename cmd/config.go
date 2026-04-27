package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  "Get and set default settings such as page size, output format, and fetch timeout.",
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		val, err := database.GetConfig(args[0])
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("config key '%s' not set", args[0])
			}
			return err
		}
		fmt.Println(val)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		validKeys := map[string]bool{
			"page_size":     true,
			"format":        true,
			"fetch_timeout": true,
		}

		if !validKeys[args[0]] {
			return fmt.Errorf("unknown config key '%s'. Valid keys: page_size, format, fetch_timeout", args[0])
		}

		if err := database.SetConfig(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Set %s = %s\n", args[0], args[1])
		return nil
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
