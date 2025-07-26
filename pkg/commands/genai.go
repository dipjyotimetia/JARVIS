package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dipjyotimetia/jarvis/pkg/engine/files"
	"github.com/dipjyotimetia/jarvis/pkg/engine/ollama"
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
