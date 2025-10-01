package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dipjyotimetia/jarvis/pkg/engine/files"
	"github.com/dipjyotimetia/jarvis/pkg/engine/llm"
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
			llmClient, err := llm.NewFromEnv(ctx)
			if err != nil {
				return fmt.Errorf("failed to create LLM client: %w", err)
			}

			// Generate tests
			prompt := buildPrompt(reader, fmt.Sprintf("Generate %s tests based on this %s spec.", language, spec))

			ct := time.Now().Format("2006-01-02-15-04-05")
			files.CheckDirectryExists(outputPath)
			outputFile, err := os.Create(fmt.Sprintf("%s/%s_output_test.md", outputPath, ct))
			if err != nil {
				s.Stop()
				return err
			}
			defer outputFile.Close()

			writer := bufio.NewWriter(outputFile)
			defer writer.Flush()

			err = llmClient.GenerateStream(ctx, prompt, func(chunk string) error {
				_, err := fmt.Fprint(writer, chunk)
				return err
			})

			if err != nil {
				s.FinalMSG = "Test generation failed: %v\n"
				s.Stop()
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
			llmClient, err := llm.NewFromEnv(ctx)
			if err != nil {
				return fmt.Errorf("failed to create LLM client: %w", err)
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

			// Generate scenarios
			promptText := buildPrompt(reader, fmt.Sprintf("Generate all possible positive and negative test scenarios in simple english for the provided %s spec file.", spec))

			err = llmClient.GenerateStream(ctx, promptText, func(chunk string) error {
				fmt.Print(chunk)
				return nil
			})

			if err != nil {
				return err
			}
			return nil
		},
	}
	setGenerateScenariosFlag(cmd)
	return cmd
}

// buildPrompt combines specs with instruction text
func buildPrompt(specs []string, instruction string) string {
	var builder strings.Builder

	for _, spec := range specs {
		builder.WriteString(spec)
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	builder.WriteString(instruction)

	return builder.String()
}
