/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package config

import (
	"github.com/spf13/cobra"

	"github.com/lorne-luo/opencommit/cmd/config/baseurl"
	"github.com/lorne-luo/opencommit/cmd/config/key"
	"github.com/lorne-luo/opencommit/cmd/config/model"
)

// ConfigCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage opencommit configuration through cli",
	Long:  `Manage opencommit configuration through cli`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("config called")
	},
}

func init() {
	ConfigCmd.AddCommand(key.KeyCmd, model.KeyCmd, baseurl.BaseurlCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
