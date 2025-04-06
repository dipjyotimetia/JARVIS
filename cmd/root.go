package cmd

import (
	"fmt"
	"os"

	conf "github.com/dipjyotimetia/jarvis/config"
	"github.com/dipjyotimetia/jarvis/pkg/commands"
	"github.com/dipjyotimetia/jarvis/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	config  conf.Config
	debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "jarvis",
	Short: "A generative AI-driven CLI for testing",
	Long:  `Jarvis is a powerful toolkit that leverages generative AI to streamline and enhance various testing activities.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set up global logging based on debug flag
		if debug {
			logger.SetGlobalLevel(logger.DebugLevel)
			logger.Debug("Debug logging enabled")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
	},
}

// Create command groups
var (
	genGroup = &cobra.Command{
		Use:   "gen",
		Short: "Generation commands",
		Long:  "Commands for generating various test artifacts",
	}

	analyzeGroup = &cobra.Command{
		Use:   "analyze",
		Short: "Analysis commands",
		Long:  "Commands for analyzing specifications and services",
	}

	toolsGroup = &cobra.Command{
		Use:   "tools",
		Short: "Utility tools",
		Long:  "Various utility tools",
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringP("author", "a", "Dipjyoti Metia", "")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output")

	// Enable bash completion
	rootCmd.CompletionOptions.DisableDefaultCmd = false

	// Add commands to groups
	genGroup.AddCommand(commands.GenerateTestModule())
	genGroup.AddCommand(commands.GenerateTestScenarios())

	analyzeGroup.AddCommand(commands.SpecAnalyzer())

	toolsGroup.AddCommand(commands.GrpcCurlGenerator())

	// Add groups to root command
	rootCmd.AddCommand(genGroup)
	rootCmd.AddCommand(analyzeGroup)
	rootCmd.AddCommand(toolsGroup)
	rootCmd.AddCommand(proxyCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(certCmd)
	rootCmd.AddCommand(commands.SetupCmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	viper.SetEnvPrefix("jarvis")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Info("🔧 Using config file: %s", viper.ConfigFileUsed())
	}
}
