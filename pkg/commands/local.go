package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dipjyotimetia/jarvis/pkg/engine/files"
	"github.com/dipjyotimetia/jarvis/pkg/engine/prompt"
	"github.com/dipjyotimetia/jarvis/pkg/engine/utils"
	"github.com/dipjyotimetia/jarvis/pkg/logger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setSpecPathFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("path", "p", "", "spec path")
}

func setGrpCurlPathFlag(cmd *cobra.Command) {
	cmd.Flags().String("proto", "", "protofile path")
	cmd.Flags().String("service", "", "service name")
	cmd.Flags().String("method", "", "method name")
}

func SpecAnalyzer() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "spec-analyzer",
		Aliases: []string{"spec"},
		Short:   "spec-analyzer is for analyzing API specification files",
		Long:    `spec-analyzer is for analyzing API specification files including Protobuf and OpenAPI`,
		RunE: func(cmd *cobra.Command, args []string) error {
			specPath, _ := cmd.Flags().GetString("path")

			// Validate required flags
			if specPath == "" {
				return fmt.Errorf("spec path is required")
			}

			specContent := prompt.PromptContent{
				ErrorMsg: "Please provide a valid spec.",
				Label:    "What spec would you like to use?",
				ItemType: "spec",
			}
			spec := prompt.SelectLanguage(specContent)

			specs, err := files.ListFiles(specPath)
			if err != nil {
				return fmt.Errorf("failed to list files: %w", err)
			}
			if len(specs) == 0 {
				return errors.New("no files found at the specified path")
			}

			fmt.Printf("Analyzing %s spec files...\n", spec)
			switch spec {
			case "protobuf":
				if err := utils.ProtoAnalyzer(specs); err != nil {
					return fmt.Errorf("failed to analyze protobuf files: %w", err)
				}
			case "openapi":
				utils.OpenApiAnalyzer(specs)
			default:
				return fmt.Errorf("unsupported spec type: %s", spec)
			}
			return nil
		},
	}
	setSpecPathFlag(cmd)
	cmd.MarkFlagRequired("path")
	return cmd
}

func GrpcCurlGenerator() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "grpc-curl",
		Aliases: []string{"grpc-curl"},
		Short:   "grpc-curl is for generating curl commands for grpc services",
		Long:    `grpc-curl is for generating curl commands for grpc services`,
		RunE: func(cmd *cobra.Command, args []string) error {
			protoFile, _ := cmd.Flags().GetString("proto")
			serviceName, _ := cmd.Flags().GetString("service")
			methodName, _ := cmd.Flags().GetString("method")

			// Validate required flags
			if protoFile == "" {
				return fmt.Errorf("proto file path is required")
			}
			if serviceName == "" {
				return fmt.Errorf("service name is required")
			}
			if methodName == "" {
				return fmt.Errorf("method name is required")
			}

			return utils.GrpCurlCommand(protoFile, serviceName, methodName)
		},
	}
	setGrpCurlPathFlag(cmd)

	// Mark required flags
	cmd.MarkFlagRequired("proto")
	cmd.MarkFlagRequired("service")
	cmd.MarkFlagRequired("method")

	return cmd
}

// SetupCmd creates an interactive setup wizard using promptui
func SetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "setup",
		Short:   "Interactive setup for Jarvis",
		Long:    `Configure Jarvis interactively with a step-by-step wizard`,
		Example: `  jarvis setup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupWizard()
		},
	}
	return cmd
}

func runSetupWizard() error {
	// Welcome message
	color.Cyan("\n🤖 Welcome to Jarvis Setup Wizard!\n")
	fmt.Println("This wizard will guide you through setting up your Jarvis configuration.")

	// Confirm proceeding with setup
	if !prompt.Confirm("Would you like to proceed with setup?", true) {
		color.Yellow("Setup cancelled. No changes were made.")
		return nil
	}

	// API Configuration
	apiKey := prompt.Password("Enter your AI API Key (leave empty for offline mode)")

	// Get default output directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	defaultDir := filepath.Join(homeDir, "jarvis-output")

	// Output directory selection
	outputDir := prompt.Input("Output directory for generated files", defaultDir, nil)

	// Ensure the directory exists
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Select programming language preference
	languageContent := prompt.PromptContent{
		ErrorMsg: "Please provide a valid language.",
		Label:    "What is your preferred programming language?",
		ItemType: "language",
	}
	language := prompt.SelectLanguage(languageContent)

	// Select framework for the language
	framework := prompt.SelectFramework(language)

	// Determine if user wants to configure proxy settings
	configureProxy := prompt.Confirm("Do you want to configure proxy settings?", false)
	var proxyPort int
	var recordMode bool

	if configureProxy {
		// Get proxy port
		proxyPort = prompt.InputNumber("Proxy port number", 8080)

		// Choose recording mode
		recordMode = prompt.Confirm("Enable recording mode?", true)
	}

	// Determine test reporting options
	reportOptions := []string{"HTML", "JSON", "JUnit XML", "Text"}
	reportFormat, err := prompt.SelectWithSearch("Select test report format", reportOptions)
	if err != nil {
		logger.Warn("Failed to select report format: %v", err)
		reportFormat = "HTML"
	}

	// Create configuration
	config := map[string]interface{}{
		"api": map[string]interface{}{
			"ai_key": apiKey,
		},
		"output": map[string]interface{}{
			"directory":     outputDir,
			"report_format": reportFormat,
		},
		"language": map[string]interface{}{
			"preferred": language,
			"framework": framework,
		},
	}

	if configureProxy {
		config["proxy"] = map[string]interface{}{
			"port":        proxyPort,
			"record_mode": recordMode,
			"passthrough": !recordMode,
		}
	}

	// Save configuration
	for key, value := range config {
		viper.Set(key, value)
	}

	configPath := "config.yaml"
	if prompt.Confirm("Save configuration to "+configPath+"?", true) {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")

		if err := viper.WriteConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file doesn't exist, create it
				if err := viper.SafeWriteConfig(); err != nil {
					return fmt.Errorf("failed to write config file: %w", err)
				}
			} else {
				return fmt.Errorf("failed to write config file: %w", err)
			}
		}

		color.Green("✅ Configuration saved to %s", configPath)
	}

	// Display summary
	color.Cyan("\n📋 Configuration Summary:")
	fmt.Printf("Language: %s\n", language)
	if framework != "" {
		fmt.Printf("Framework: %s\n", framework)
	}
	fmt.Printf("Output Directory: %s\n", outputDir)
	fmt.Printf("Report Format: %s\n", reportFormat)

	if configureProxy {
		fmt.Printf("Proxy Port: %d\n", proxyPort)
		fmt.Printf("Recording Mode: %v\n", recordMode)
	}

	color.Green("\n🎉 Setup completed successfully!")
	return nil
}
