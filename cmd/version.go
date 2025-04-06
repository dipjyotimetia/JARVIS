package cmd

import (
	"fmt"
	"runtime"

	"github.com/dipjyotimetia/jarvis/pkg/github"
	"github.com/dipjyotimetia/jarvis/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	// These variables are set during build using -ldflags
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Display detailed version information about the Jarvis CLI, including version number, commit hash, and build date.`,
	Example: `  # Display version information
  jarvis version`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("🚀 Jarvis CLI version information:")
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Commit:     %s\n", Commit)
		fmt.Printf("Built on:   %s\n", BuildDate)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Jarvis CLI to the latest version",
	Long:  `Check for a newer version of Jarvis CLI and automatically update to the latest release if available.`,
	Example: `  # Check for updates and apply if available
  jarvis update`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Checking for updates to Jarvis CLI...")

		err := github.SelfUpdate(Version)
		if err != nil {
			logger.Error("%s", fmt.Sprintf("Update failed: %s", err))
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
}
