package cmd

import (
	"fmt"
	"os"

	conf "github.com/dipjyotimetia/jarvis/config"
	"github.com/dipjyotimetia/jarvis/pkg/commands"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	config  conf.Config
)

var rootCmd = &cobra.Command{
	Use:   "jarvis",
	Short: "A generative AI-driven CLI for testing",
	Long:  `Jarvis is a powerful toolkit that leverages generative AI to streamline and enhance various testing activities.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringP("author", "a", "Dipjyoti Metia", "")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	rootCmd.AddCommand(commands.GenerateTestModule())
	rootCmd.AddCommand(commands.GenerateTestScenarios())
	rootCmd.AddCommand(commands.SpecAnalyzer())
	rootCmd.AddCommand(commands.GrpcCurlGenerator())
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
		fmt.Println("🔧 Using config file:", viper.ConfigFileUsed())
	}
}
