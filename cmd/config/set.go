/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ValidConfigKeys defines all valid configuration keys and their types
var ValidConfigKeys = map[string]string{
	// [api]
	"api.key":          "string",
	"api.model":        "string",
	"api.baseurl":      "string",
	"api.last_success": "int",
	// [api2] — secondary provider for fallback
	"api2.key":     "string",
	"api2.model":   "string",
	"api2.baseurl": "string",
	// [commit]
	"commit.language":   "string",
	"commit.max_length":     "int",
	"commit.max_diff_lines": "int",
	// [behavior]
	"behavior.stage_all":   "bool",
	"behavior.auto_select": "bool",
	"behavior.no_confirm":  "bool",
	"behavior.quiet":       "bool",
	"behavior.push":        "bool",
	"behavior.dry_run":     "bool",
	"behavior.show_diff":   "bool",
	"behavior.no_verify":   "bool",
}

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

[api]
  api.key             - AI provider API key
  api.model           - AI provider model name (default: gpt-3.5-turbo)
  api.baseurl         - Custom base URL for AI provider API
  api.last_success    - Provider used last successfully: 1 or 2 (managed automatically)

[api2] (optional secondary provider — used when api1 fails)
  api2.key            - Secondary AI provider API key
  api2.model          - Secondary AI provider model name
  api2.baseurl        - Secondary custom base URL for AI provider API

[commit]
  commit.language     - Language for commit messages (default: english)
  commit.max_length     - Maximum length of commit message (default: 72)
  commit.max_diff_lines - Truncate per-file diff to N lines to save tokens (default: 500, 0 disables)

[behavior]
  behavior.stage_all   - Stage all changes in tracked files (default: false)
  behavior.auto_select - Let AI select files and generate commit message (default: false)
  behavior.no_confirm  - Skip confirmation prompt (default: false)
  behavior.quiet       - Suppress output (default: false)
  behavior.push        - Push committed changes to remote (default: false)
  behavior.dry_run     - Run without making changes (default: false)
  behavior.show_diff   - Show diff before committing (default: false)
  behavior.no_verify   - Skip git commit-msg hook verification (default: false)

Example:
  opencommit config set commit.language korean
  opencommit config set commit.max_length 100
  opencommit config set behavior.push true`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		keyType, valid := ValidConfigKeys[key]
		if !valid {
			fmt.Printf("Error: unknown config key '%s'\n", key)
			fmt.Println("Run 'opencommit config set --help' to see available keys")
			os.Exit(1)
		}

		var finalValue interface{}
		switch keyType {
		case "int":
			intVal, err := strconv.Atoi(value)
			if err != nil {
				fmt.Printf("Error: value '%s' is not a valid integer\n", value)
				os.Exit(1)
			}
			finalValue = intVal
		case "bool":
			boolVal, err := strconv.ParseBool(value)
			if err != nil {
				fmt.Printf("Error: value '%s' is not a valid boolean (use true/false)\n", value)
				os.Exit(1)
			}
			finalValue = boolVal
		default:
			finalValue = value
		}

		viper.Set(key, finalValue)
		if err := viper.WriteConfig(); err != nil {
			fmt.Printf("Error: failed to write config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set %s = %v\n", key, finalValue)
	},
}

func init() {
	ConfigCmd.AddCommand(setCmd)
}
