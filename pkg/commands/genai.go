package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dipjyotimetia/jarvis/pkg/engine/files"
	"github.com/dipjyotimetia/jarvis/pkg/engine/ollama"
	"github.com/dipjyotimetia/jarvis/pkg/engine/pact"
	"github.com/dipjyotimetia/jarvis/pkg/engine/prompt"
	"github.com/spf13/cobra"
)

func setGenerateTestFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("path", "p", "", "spec path")
	cmd.Flags().StringP("output", "o", "", "output path")
}

func setGenerateScenariosFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("path", "p", "", "spec path")
}

func setGenerateContractsFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("path", "p", "", "spec path")
	cmd.Flags().StringP("output", "o", "./contracts", "output path")
	cmd.Flags().StringP("consumer", "c", "", "consumer name")
	cmd.Flags().StringP("provider", "r", "", "provider name")
	cmd.Flags().StringP("language", "l", "", "target language for test code")
	cmd.Flags().StringP("framework", "f", "", "target framework for test code")
	cmd.Flags().Bool("examples", false, "include test code examples")
}

func GenerateTestModule() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate-test",
		Aliases: []string{"test"},
		Short:   "generate-test is for generating test cases.",
		Long:    `generate-test is for generating test cases from the provided spec files`,
		RunE: func(cmd *cobra.Command, args []string) error {
			specPath, _ := cmd.Flags().GetString("path")
			outputPath, _ := cmd.Flags().GetString("output")

			s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
			s.Color("green")
			s.Suffix = " Generating Tests..."
			s.FinalMSG = "Tests Generated Successfully!\n"

			languageContent := prompt.PromptContent{
				ErrorMsg: "Please provide a valid language.",
				Label:    "What programming lanaguage would you like to use?",
				ItemType: "language",
			}
			language := prompt.SelectLanguage(languageContent)

			specContent := prompt.PromptContent{
				ErrorMsg: "Please provide a valid spec.",
				Label:    "What spec would you like to use?",
				ItemType: "spec",
			}
			spec := prompt.SelectLanguage(specContent)

			file, err := files.ListFiles(specPath)
			if err != nil {
				return fmt.Errorf("failed to identify spec types: %w", err)
			}
			if len(file) == 0 {
				return errors.New("no files found")
			}

			reader, err := files.ReadFile(file[0])
			if err != nil {
				return fmt.Errorf("failed to read spec file: %w", err)
			}

			s.Start()
			ctx := context.Background()
			ai, err := ollama.New(ctx)
			if err != nil {
				return fmt.Errorf("failed to create Ollama engine: %w", err)
			}

			err = ai.GenerateTextStreamWriter(ctx, reader, language, spec, outputPath)
			if err != nil {
				s.FinalMSG = "Test generation failed: %v\n"
				return err
			}
			s.Stop()
			return nil
		},
	}
	setGenerateTestFlag(cmd)
	return cmd
}

func GenerateTestScenarios() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate-scenarios",
		Aliases: []string{"scenarios"},
		Short:   "generate-scenarios is for generating test scenarios.",
		Long:    `generate-scenarios is for generating test scenarios from the provided spec files`,
		RunE: func(cmd *cobra.Command, args []string) error {
			specPath, _ := cmd.Flags().GetString("path")

			specContent := prompt.PromptContent{
				ErrorMsg: "Please provide a valid spec.",
				Label:    "What spec would you like to use?",
				ItemType: "spec",
			}

			spec := prompt.SelectLanguage(specContent)

			ctx := context.Background()
			ai, err := ollama.New(ctx)
			if err != nil {
				return fmt.Errorf("failed to create Ollama engine: %w", err)
			}

			file, err := files.ListFiles(specPath)
			if err != nil {
				return fmt.Errorf("failed to identify spec types: %w", err)
			}
			if len(file) == 0 {
				return errors.New("no files found")
			}

			reader, err := files.ReadFile(file[0])
			if err != nil {
				return fmt.Errorf("failed to read spec file: %w", err)
			}

			err = ai.GenerateTextStream(ctx, reader, spec)
			if err != nil {
				return err
			}
			return nil
		},
	}
	setGenerateScenariosFlag(cmd)
	return cmd
}

func GenerateContractsModule() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate-contracts",
		Aliases: []string{"contracts", "pact"},
		Short:   "generate-contracts is for generating Pact contracts.",
		Long:    `generate-contracts is for generating Pact contract files from OpenAPI specifications using AI`,
		Example: `  # Generate contracts from OpenAPI spec
  jarvis generate-contracts --path="specs/openapi/v3.0/my_api.yaml" --consumer="web-app" --provider="api-service"
  
  # Generate contracts with test code examples
  jarvis generate-contracts --path="specs/openapi" --consumer="mobile-app" --provider="backend-api" --language="javascript" --framework="jest" --examples
  
  # Generate contracts to specific output directory
  jarvis generate-contracts --path="specs/openapi/api.yaml" --output="./pact-contracts" --consumer="frontend" --provider="backend"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			specPath, _ := cmd.Flags().GetString("path")
			outputPath, _ := cmd.Flags().GetString("output")
			consumerName, _ := cmd.Flags().GetString("consumer")
			providerName, _ := cmd.Flags().GetString("provider")
			language, _ := cmd.Flags().GetString("language")
			framework, _ := cmd.Flags().GetString("framework")
			includeExamples, _ := cmd.Flags().GetBool("examples")

			if specPath == "" {
				return errors.New("spec path is required")
			}

			if consumerName == "" {
				return errors.New("consumer name is required")
			}

			if providerName == "" {
				return errors.New("provider name is required")
			}

			s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
			s.Color("green")
			s.Suffix = " Generating Pact Contracts..."
			s.FinalMSG = "Pact Contracts Generated Successfully!\n"

			ctx := context.Background()

			// Create Pact generation configuration
			config := &pact.GenerationConfig{
				ConsumerName:    consumerName,
				ProviderName:    providerName,
				OutputPath:      outputPath,
				SpecVersion:     "3.0.0",
				IncludeExamples: includeExamples,
				Language:        language,
				Framework:       framework,
			}

			// Create Pact generator
			generator, err := pact.NewGenerator(ctx, config)
			if err != nil {
				return fmt.Errorf("failed to create Pact generator: %w", err)
			}
			defer generator.Close()

			s.Start()

			// Generate contracts from OpenAPI spec
			result, err := generator.GenerateFromOpenAPI(ctx, specPath)
			if err != nil {
				s.FinalMSG = fmt.Sprintf("Contract generation failed: %v\n", err)
				s.Stop()
				return err
			}

			s.Stop()

			// Validate the generated contract with detailed feedback
			detailedValidation := generator.ValidateContractDetailed(result.Contract, false)
			if !detailedValidation.Valid {
				fmt.Printf("⚠️  Contract validation issues found:\n")
				for _, err := range detailedValidation.Errors {
					fmt.Printf("   ❌ %s: %s\n", err.Location, err.Message)
					if err.Suggestion != "" {
						fmt.Printf("      💡 %s\n", err.Suggestion)
					}
				}
				for _, warning := range detailedValidation.Warnings {
					fmt.Printf("   ⚠️  %s: %s\n", warning.Location, warning.Message)
					if warning.Suggestion != "" {
						fmt.Printf("      💡 %s\n", warning.Suggestion)
					}
				}
			} else {
				fmt.Printf("✅ Contract validation passed\n")
			}

			// Display suggestions for improvement
			if len(detailedValidation.Suggestions) > 0 {
				fmt.Printf("\n💡 Suggestions for improvement:\n")
				for _, suggestion := range detailedValidation.Suggestions {
					fmt.Printf("   • %s\n", suggestion.Message)
					if suggestion.Example != "" {
						fmt.Printf("     Example: %s\n", suggestion.Example)
					}
				}
			}

			// Display results
			fmt.Printf("📄 Contract file generated: %s\n", result.FilePath)
			fmt.Printf("🔗 Interactions generated: %d\n", result.Interactions)
			fmt.Printf("👥 Consumer: %s\n", result.Contract.Consumer.Name)
			fmt.Printf("🏪 Provider: %s\n", result.Contract.Provider.Name)

			if result.TestCode != "" {
				testFilePath := fmt.Sprintf("%s/%s_%s_test.%s", 
					outputPath, 
					strings.ToLower(consumerName),
					strings.ToLower(providerName),
					getFileExtension(language))
				
				if err := os.WriteFile(testFilePath, []byte(result.TestCode), 0644); err != nil {
					fmt.Printf("⚠️  Failed to save test code: %v\n", err)
				} else {
					fmt.Printf("🧪 Test code generated: %s\n", testFilePath)
				}
			}

			return nil
		},
	}
	setGenerateContractsFlag(cmd)
	return cmd
}

// getFileExtension returns the file extension for a given language
func getFileExtension(language string) string {
	switch strings.ToLower(language) {
	case "javascript", "js":
		return "js"
	case "typescript", "ts":
		return "ts"
	case "python", "py":
		return "py"
	case "java":
		return "java"
	case "go", "golang":
		return "go"
	case "ruby", "rb":
		return "rb"
	case "php":
		return "php"
	case "csharp", "c#":
		return "cs"
	default:
		return "txt"
	}
}
