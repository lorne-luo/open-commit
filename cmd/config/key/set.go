/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package key

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// printConfigFileLocation prints the path where the configuration is saved
func printConfigFileLocation() {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = os.Getenv("HOME")
		}
		configFile = filepath.Join(homeDir, ".config", "opencommit", "config.toml")
	}
	fmt.Printf("Configuration saved to: %s\n", configFile)
}

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set {api_key}",
	Short: "Set AI API key",
	Long:  `Set AI API key`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := args[0]
		viper.Set("api.key", apiKey)
		cobra.CheckErr(viper.WriteConfig())
		printConfigFileLocation()
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
