// Package config provides the config for CrashDragon
package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
)

// GetConfig loads default values and overwrites them by the ones in a file, or creates a file with them if there is no file
func GetConfig() error {

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/crashdragon/")
	viper.AddConfigPath("$HOME/.crashdragon")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		// Set default values if there is no config
		viper.SetDefault("DB.Connection", "host=localhost user=crashdragon dbname=crashdragon sslmode=disable")
		viper.SetDefault("Web.UseSocket", false)
		viper.SetDefault("Web.BindAddress", ":8080")
		viper.SetDefault("Web.BindSocket", "/var/run/crashdragon/crashdragon.sock")
		viper.SetDefault("Directory.Content", "../share/crashdragon/files")
		viper.SetDefault("Directory.Templates", "./web/templates")
		viper.SetDefault("Directory.Assets", "./web/assets")
		viper.SetDefault("Symbolicator.Executable", "./minidump_stackwalk")
		viper.SetDefault("Symbolicator.TrimModuleNames", true)
		viper.SetDefault("Housekeeping.ReportRetentionTime", "2190h") // Around 3 months (duration only supports times in hours and down due to irregular length of days/months/years)
		viper.SetDefault("Slack.webhook", "https://hooks.slack.com/services/")

		// Get the path of the Go executable
		exePath, err := os.Executable()
		if err != nil {
			log.Fatalf("Failed to get the executable path: %+v", err)
		}

		// Get the current folder (directory) of the executable
		configFile := filepath.Dir(exePath) + "/config.toml"
		// Write the configuration to file
		err = viper.WriteConfigAs(configFile) // Specify the filename and extension
		if err != nil {
			// Handle the error if writing fails
			log.Fatalf("Failed to write the config file: %+v", err)
		} else {
			log.Fatalf("Sample config file has been created %+v, please, review it, and then relaunch.", configFile)
		}
	}

	return viper.WriteConfig()
}
