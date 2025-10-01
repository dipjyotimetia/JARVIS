package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dipjyotimetia/jarvis/internal/db"
	"github.com/dipjyotimetia/jarvis/pkg/engine/analyzer"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func setAnalyzeFailuresFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("database", "d", "traffic_inspector.db", "Path to traffic database")
	cmd.Flags().StringP("endpoint", "e", "", "Filter by specific endpoint (optional)")
	cmd.Flags().IntP("status", "s", 0, "Filter by status code (0 for all failures)")
	cmd.Flags().IntP("limit", "l", 10, "Maximum number of failures to analyze")
	cmd.Flags().DurationP("time-window", "t", 24*time.Hour, "Analyze failures within this time window")
	cmd.Flags().StringP("output", "o", "", "Output file for detailed report (JSON)")
	cmd.Flags().Bool("stats-only", false, "Show only statistics without detailed analysis")
}

func AnalyzeFailuresCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "analyze-failures",
		Aliases: []string{"failures", "analyze-errors"},
		Short:   "Analyze API failures with AI-powered diagnosis",
		Long: `Analyze failed API requests from recorded traffic and provide AI-powered diagnosis.

This command analyzes:
- Root causes of failures
- Error patterns and categories
- Actionable suggestions for fixes
- Related issues and documentation
- Prevention tips

Examples:
  # Analyze recent failures
  jarvis analyze analyze-failures

  # Analyze specific endpoint failures
  jarvis analyze analyze-failures --endpoint /api/users

  # Analyze only 500 errors
  jarvis analyze analyze-failures --status 500

  # Show statistics only
  jarvis analyze analyze-failures --stats-only

  # Analyze last hour's failures
  jarvis analyze analyze-failures --time-window 1h

  # Save detailed report to file
  jarvis analyze analyze-failures --output failures-report.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, _ := cmd.Flags().GetString("database")
			endpoint, _ := cmd.Flags().GetString("endpoint")
			statusCode, _ := cmd.Flags().GetInt("status")
			limit, _ := cmd.Flags().GetInt("limit")
			timeWindow, _ := cmd.Flags().GetDuration("time-window")
			outputFile, _ := cmd.Flags().GetString("output")
			statsOnly, _ := cmd.Flags().GetBool("stats-only")

			// Open database
			database, _, err := db.Initialize(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer database.Close()

			ctx := context.Background()

			// Create analyzer
			failureAnalyzer, err := analyzer.New(ctx, database)
			if err != nil {
				return fmt.Errorf("failed to create analyzer: %w", err)
			}

			// Show statistics first
			fmt.Println("📊 Failure Statistics")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

			stats, err := failureAnalyzer.GetFailureStatistics(ctx, timeWindow)
			if err != nil {
				return fmt.Errorf("failed to get statistics: %w", err)
			}

			printFailureStatistics(stats)

			if statsOnly {
				return nil
			}

			// Perform detailed analysis
			fmt.Println("\n🔍 Analyzing Failures...")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

			s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
			s.Color("yellow")
			s.Suffix = " Running AI-powered failure analysis..."
			s.Start()

			opts := analyzer.AnalysisOptions{
				TimeWindow: timeWindow,
				StatusCode: statusCode,
				Endpoint:   endpoint,
				Limit:      limit,
			}

			reports, err := failureAnalyzer.AnalyzeFailures(ctx, opts)
			if err != nil {
				s.Stop()
				return fmt.Errorf("failed to analyze failures: %w", err)
			}

			s.Stop()

			if len(reports) == 0 {
				fmt.Println("\n✅ No failures found in the specified time window!")
				return nil
			}

			// Print detailed analysis
			printFailureReports(reports)

			// Save to file if requested
			if outputFile != "" {
				jsonData, err := json.MarshalIndent(reports, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal reports: %w", err)
				}

				if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}

				fmt.Printf("\n💾 Detailed report saved to: %s\n", outputFile)
			}

			return nil
		},
	}

	setAnalyzeFailuresFlags(cmd)
	return cmd
}

func printFailureStatistics(stats analyzer.FailureStats) {
	if stats.Total == 0 {
		fmt.Println("✅ No failures found!")
		return
	}

	fmt.Printf("\nTotal Failures: %d\n\n", stats.Total)

	// Create table
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Status Code", "Count", "Unique Endpoints", "Percentage")

	for statusCode, codeStats := range stats.ByStatusCode {
		percentage := float64(codeStats.Count) / float64(stats.Total) * 100
		table.Append(
			fmt.Sprintf("%d", statusCode),
			fmt.Sprintf("%d", codeStats.Count),
			fmt.Sprintf("%d", codeStats.UniqueEndpoints),
			fmt.Sprintf("%.1f%%", percentage),
		)
	}

	table.Render()
}

func printFailureReports(reports []analyzer.FailureReport) {
	fmt.Printf("\n📋 Analyzed %d Failures\n", len(reports))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, report := range reports {
		fmt.Printf("\n%d. %s %s\n", i+1, report.Method, report.URL)
		fmt.Println("   ────────────────────────────────────────")

		// Status and timing
		statusColor := color.New(color.FgRed, color.Bold)
		if report.StatusCode < 500 {
			statusColor = color.New(color.FgYellow, color.Bold)
		}
		statusColor.Printf("   Status: %d", report.StatusCode)
		fmt.Printf(" | Duration: %dms | Time: %s\n",
			report.Duration,
			report.Timestamp.Format("2006-01-02 15:04:05"))

		// Analysis
		analysis := report.Analysis

		// Root cause
		fmt.Printf("\n   🎯 Root Cause:\n")
		fmt.Printf("      %s\n", analysis.RootCause)

		// Category and severity
		fmt.Printf("\n   📂 Category: %s | ", analysis.ErrorCategory)
		severityColor := getSeverityColor(analysis.Severity)
		severityColor.Printf("Severity: %s", analysis.Severity)
		if analysis.FixEstimate != "" {
			fmt.Printf(" | Estimate: %s", analysis.FixEstimate)
		}
		fmt.Println()

		// Suggestions
		if len(analysis.Suggestions) > 0 {
			fmt.Printf("\n   💡 Suggestions:\n")
			for _, suggestion := range analysis.Suggestions {
				fmt.Printf("      • %s\n", suggestion)
			}
		}

		// Possible fixes
		if len(analysis.PossibleFixes) > 0 {
			fmt.Printf("\n   🔧 Possible Fixes:\n")
			for _, fix := range analysis.PossibleFixes {
				fmt.Printf("      • %s\n", fix)
			}
		}

		// Reproduction steps
		if len(analysis.ReproSteps) > 0 {
			fmt.Printf("\n   🔄 Reproduction Steps:\n")
			for j, step := range analysis.ReproSteps {
				fmt.Printf("      %d. %s\n", j+1, step)
			}
		}

		// Prevention tips
		if len(analysis.PreventionTips) > 0 {
			fmt.Printf("\n   🛡️  Prevention:\n")
			for _, tip := range analysis.PreventionTips {
				fmt.Printf("      • %s\n", tip)
			}
		}

		// Related issues
		if len(analysis.RelatedIssues) > 0 {
			fmt.Printf("\n   🔗 Related Issues:\n")
			for _, issue := range analysis.RelatedIssues {
				fmt.Printf("      • %s\n", issue)
			}
		}

		// Documentation
		if len(analysis.Documentation) > 0 {
			fmt.Printf("\n   📚 Documentation:\n")
			for _, doc := range analysis.Documentation {
				fmt.Printf("      • %s\n", doc)
			}
		}

		fmt.Println()
	}
}

func getSeverityColor(severity string) *color.Color {
	switch severity {
	case "Critical":
		return color.New(color.FgRed, color.Bold)
	case "High":
		return color.New(color.FgRed)
	case "Medium":
		return color.New(color.FgYellow)
	case "Low":
		return color.New(color.FgGreen)
	default:
		return color.New(color.FgWhite)
	}
}

func AnalyzeEndpointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze-endpoint",
		Short: "Analyze failures for a specific endpoint",
		Long: `Analyze all failures for a specific API endpoint.

Examples:
  # Analyze failures for user endpoint
  jarvis analyze analyze-endpoint /api/users --limit 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("endpoint path required")
			}

			endpoint := args[0]
			dbPath, _ := cmd.Flags().GetString("database")
			limit, _ := cmd.Flags().GetInt("limit")

			// Open database
			database, _, err := db.Initialize(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer database.Close()

			ctx := context.Background()

			// Create analyzer
			failureAnalyzer, err := analyzer.New(ctx, database)
			if err != nil {
				return fmt.Errorf("failed to create analyzer: %w", err)
			}

			fmt.Printf("🔍 Analyzing failures for: %s\n", endpoint)

			s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
			s.Color("yellow")
			s.Suffix = " Analyzing endpoint failures..."
			s.Start()

			reports, err := failureAnalyzer.AnalyzeEndpoint(ctx, endpoint, limit)
			if err != nil {
				s.Stop()
				return fmt.Errorf("failed to analyze endpoint: %w", err)
			}

			s.Stop()

			if len(reports) == 0 {
				fmt.Println("\n✅ No failures found for this endpoint!")
				return nil
			}

			printFailureReports(reports)

			return nil
		},
	}

	cmd.Flags().StringP("database", "d", "traffic_inspector.db", "Path to traffic database")
	cmd.Flags().IntP("limit", "l", 10, "Maximum number of failures to analyze")

	return cmd
}
