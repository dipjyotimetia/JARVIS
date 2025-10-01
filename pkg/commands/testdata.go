package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dipjyotimetia/jarvis/pkg/engine/testdata"
	"github.com/spf13/cobra"
)

func setGenerateTestDataFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("spec", "s", "", "Path to OpenAPI specification file (required)")
	cmd.Flags().StringP("output", "o", "./testdata", "Output directory for generated test data")
	cmd.Flags().IntP("count", "c", 5, "Number of test cases to generate per endpoint")
	cmd.Flags().Bool("valid", true, "Generate valid test cases")
	cmd.Flags().Bool("invalid", true, "Generate invalid/boundary test cases")
	cmd.Flags().StringP("format", "f", "json", "Output format (json, yaml)")
	cmd.Flags().String("locale", "en-US", "Data locale for generation (en-US, en-GB, etc.)")
	cmd.MarkFlagRequired("spec")
}

func GenerateTestDataCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate-testdata",
		Aliases: []string{"testdata", "gen-data"},
		Short:   "Generate realistic test data from OpenAPI specifications",
		Long: `Generate AI-powered realistic test data from OpenAPI specifications.

This command analyzes your API specification and generates:
- Valid test cases that should succeed
- Invalid/boundary test cases for validation testing
- Realistic data matching your schema constraints
- Edge cases and corner scenarios

Examples:
  # Generate test data from OpenAPI spec
  jarvis gen generate-testdata --spec api.yaml

  # Generate only valid test cases
  jarvis gen generate-testdata --spec api.yaml --invalid=false

  # Generate 10 test cases per endpoint
  jarvis gen generate-testdata --spec api.yaml --count 10

  # Use UK locale for data generation
  jarvis gen generate-testdata --spec api.yaml --locale en-GB`,
		RunE: func(cmd *cobra.Command, args []string) error {
			specPath, _ := cmd.Flags().GetString("spec")
			outputDir, _ := cmd.Flags().GetString("output")
			count, _ := cmd.Flags().GetInt("count")
			includeValid, _ := cmd.Flags().GetBool("valid")
			includeInvalid, _ := cmd.Flags().GetBool("invalid")
			format, _ := cmd.Flags().GetString("format")
			locale, _ := cmd.Flags().GetString("locale")

			// Validate inputs
			if _, err := os.Stat(specPath); os.IsNotExist(err) {
				return fmt.Errorf("spec file not found: %s", specPath)
			}

			// Create output directory
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Setup spinner
			s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
			s.Color("green")
			s.Suffix = " Generating test data with AI..."
			s.Start()

			// Create test data generator
			ctx := context.Background()
			generator, err := testdata.New(ctx)
			if err != nil {
				s.Stop()
				return fmt.Errorf("failed to create test data generator: %w", err)
			}

			// Generate test data
			opts := testdata.GenerationOptions{
				Count:          count,
				IncludeValid:   includeValid,
				IncludeInvalid: includeInvalid,
				Format:         format,
				Locale:         locale,
			}

			results, err := generator.GenerateFromOpenAPI(ctx, specPath, opts)
			if err != nil {
				s.Stop()
				return fmt.Errorf("failed to generate test data: %w", err)
			}

			s.Stop()

			// Write results to files
			timestamp := time.Now().Format("2006-01-02-15-04-05")
			outputFile := fmt.Sprintf("%s/testdata_%s.json", outputDir, timestamp)

			jsonData, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal results: %w", err)
			}

			if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}

			// Print summary
			fmt.Println("\n✅ Test data generation completed!")
			fmt.Printf("📊 Generated test data for %d endpoints\n", len(results))
			fmt.Printf("💾 Output saved to: %s\n", outputFile)

			// Print statistics
			totalTestCases := 0
			for _, result := range results {
				totalTestCases += len(result.TestCases)
			}
			fmt.Printf("🧪 Total test cases: %d\n", totalTestCases)

			return nil
		},
	}

	setGenerateTestDataFlags(cmd)
	return cmd
}

func GenerateFromSchemaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-from-schema",
		Short: "Generate test data from a JSON schema",
		Long: `Generate test data from a standalone JSON schema file.

Examples:
  # Generate from schema file
  jarvis gen generate-from-schema --schema user.schema.json --count 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			schemaPath, _ := cmd.Flags().GetString("schema")
			count, _ := cmd.Flags().GetInt("count")
			includeValid, _ := cmd.Flags().GetBool("valid")
			includeInvalid, _ := cmd.Flags().GetBool("invalid")

			// Read schema file
			schemaData, err := os.ReadFile(schemaPath)
			if err != nil {
				return fmt.Errorf("failed to read schema file: %w", err)
			}

			// Setup spinner
			s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
			s.Color("green")
			s.Suffix = " Generating test data from schema..."
			s.Start()

			// Create generator
			ctx := context.Background()
			generator, err := testdata.New(ctx)
			if err != nil {
				s.Stop()
				return fmt.Errorf("failed to create generator: %w", err)
			}

			// Generate test data
			opts := testdata.GenerationOptions{
				Count:          count,
				IncludeValid:   includeValid,
				IncludeInvalid: includeInvalid,
			}

			results, err := generator.GenerateForSchema(ctx, string(schemaData), opts)
			if err != nil {
				s.Stop()
				return fmt.Errorf("failed to generate test data: %w", err)
			}

			s.Stop()

			// Output results
			jsonData, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(jsonData))

			return nil
		},
	}

	cmd.Flags().StringP("schema", "s", "", "Path to JSON schema file (required)")
	cmd.Flags().IntP("count", "c", 5, "Number of test cases to generate")
	cmd.Flags().Bool("valid", true, "Generate valid test cases")
	cmd.Flags().Bool("invalid", true, "Generate invalid test cases")
	cmd.MarkFlagRequired("schema")

	return cmd
}
