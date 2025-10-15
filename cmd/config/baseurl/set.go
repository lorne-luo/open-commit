/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package baseurl

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
	Use:   "set {base_url}",
	Short: "Set custom base URL for AI API",
	Long:  `Set custom base URL for AI API`,
	Run: func(cmd *cobra.Command, args []string) {
		baseUrl := args[0]
		viper.Set("api.baseurl", baseUrl)
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
