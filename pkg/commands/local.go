package commands

import (
	"errors"
	"fmt"

	"github.com/dipjyotimetia/jarvis/pkg/engine/files"
	"github.com/dipjyotimetia/jarvis/pkg/engine/prompt"
	"github.com/dipjyotimetia/jarvis/pkg/engine/utils"
	"github.com/spf13/cobra"
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
