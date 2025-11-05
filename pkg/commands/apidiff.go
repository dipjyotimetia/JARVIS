package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
)

// APIDiffCommand creates a command to compare two OpenAPI specifications
func APIDiffCommand() *cobra.Command {
	var oldSpecPath, newSpecPath, outputPath string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two OpenAPI specifications and detect breaking changes",
		Long: `Compare two versions of an OpenAPI specification and detect:
- Breaking changes (removed endpoints, changed required fields, etc.)
- New endpoints and features
- Deprecated functionality
- Schema changes

Examples:
  # Compare two API specs
  jarvis analyze diff --old api-v1.yaml --new api-v2.yaml

  # Generate JSON output for CI/CD
  jarvis analyze diff --old api-v1.yaml --new api-v2.yaml --json

  # Save diff to file
  jarvis analyze diff --old api-v1.yaml --new api-v2.yaml --output diff-report.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if oldSpecPath == "" || newSpecPath == "" {
				return fmt.Errorf("both --old and --new spec paths are required")
			}

			// Load old spec
			oldSpec, err := loadOpenAPISpec(oldSpecPath)
			if err != nil {
				return fmt.Errorf("failed to load old spec: %w", err)
			}

			// Load new spec
			newSpec, err := loadOpenAPISpec(newSpecPath)
			if err != nil {
				return fmt.Errorf("failed to load new spec: %w", err)
			}

			// Compare specs
			diff := compareSpecs(oldSpec, newSpec)

			// Output results
			if jsonOutput || outputPath != "" {
				data, err := json.MarshalIndent(diff, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal diff: %w", err)
				}

				if outputPath != "" {
					if err := os.WriteFile(outputPath, data, 0644); err != nil {
						return fmt.Errorf("failed to write output file: %w", err)
					}
					fmt.Printf("✅ Diff report saved to: %s\n", outputPath)
				} else {
					fmt.Println(string(data))
				}
			} else {
				printHumanReadableDiff(diff)
			}

			// Exit with error if breaking changes detected
			if len(diff.BreakingChanges) > 0 {
				fmt.Printf("\n❌ Found %d breaking change(s)\n", len(diff.BreakingChanges))
				return fmt.Errorf("breaking changes detected")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&oldSpecPath, "old", "", "Path to old OpenAPI specification")
	cmd.Flags().StringVar(&newSpecPath, "new", "", "Path to new OpenAPI specification")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output file path (optional)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

// SpecDiff represents the difference between two API specifications
type SpecDiff struct {
	BreakingChanges []Change `json:"breaking_changes"`
	NewEndpoints    []Change `json:"new_endpoints"`
	RemovedEndpoints []Change `json:"removed_endpoints"`
	ModifiedEndpoints []Change `json:"modified_endpoints"`
	DeprecatedItems []Change `json:"deprecated_items"`
	Summary         DiffSummary `json:"summary"`
}

// Change represents a single change in the API
type Change struct {
	Type        string `json:"type"`
	Path        string `json:"path"`
	Method      string `json:"method,omitempty"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	OldValue    string `json:"old_value,omitempty"`
	NewValue    string `json:"new_value,omitempty"`
}

// DiffSummary provides a summary of changes
type DiffSummary struct {
	TotalBreakingChanges int `json:"total_breaking_changes"`
	TotalNewEndpoints    int `json:"total_new_endpoints"`
	TotalRemovedEndpoints int `json:"total_removed_endpoints"`
	TotalModifications   int `json:"total_modifications"`
	Compatible           bool `json:"compatible"`
}

func loadOpenAPISpec(path string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	if err := doc.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("spec validation failed: %w", err)
	}

	return doc, nil
}

func compareSpecs(oldSpec, newSpec *openapi3.T) *SpecDiff {
	diff := &SpecDiff{
		BreakingChanges:   []Change{},
		NewEndpoints:      []Change{},
		RemovedEndpoints:  []Change{},
		ModifiedEndpoints: []Change{},
		DeprecatedItems:   []Change{},
	}

	// Track all paths
	oldPaths := make(map[string]bool)
	newPaths := make(map[string]bool)

	// Collect old paths
	for path := range oldSpec.Paths.Map() {
		oldPaths[path] = true
	}

	// Collect new paths
	for path := range newSpec.Paths.Map() {
		newPaths[path] = true
	}

	// Find removed endpoints (breaking change)
	for path := range oldPaths {
		if !newPaths[path] {
			oldPathItem := oldSpec.Paths.Find(path)
			for method := range getOperations(oldPathItem) {
				diff.BreakingChanges = append(diff.BreakingChanges, Change{
					Type:        "removed_endpoint",
					Path:        path,
					Method:      method,
					Description: fmt.Sprintf("Endpoint %s %s has been removed", method, path),
					Severity:    "high",
				})
				diff.RemovedEndpoints = append(diff.RemovedEndpoints, Change{
					Type:   "removed_endpoint",
					Path:   path,
					Method: method,
				})
			}
		}
	}

	// Find new endpoints
	for path := range newPaths {
		if !oldPaths[path] {
			newPathItem := newSpec.Paths.Find(path)
			for method := range getOperations(newPathItem) {
				diff.NewEndpoints = append(diff.NewEndpoints, Change{
					Type:        "new_endpoint",
					Path:        path,
					Method:      method,
					Description: fmt.Sprintf("New endpoint %s %s added", method, path),
					Severity:    "info",
				})
			}
		}
	}

	// Compare existing endpoints
	for path := range oldPaths {
		if newPaths[path] {
			oldPathItem := oldSpec.Paths.Find(path)
			newPathItem := newSpec.Paths.Find(path)
			comparePathItems(path, oldPathItem, newPathItem, diff)
		}
	}

	// Calculate summary
	diff.Summary = DiffSummary{
		TotalBreakingChanges:  len(diff.BreakingChanges),
		TotalNewEndpoints:     len(diff.NewEndpoints),
		TotalRemovedEndpoints: len(diff.RemovedEndpoints),
		TotalModifications:    len(diff.ModifiedEndpoints),
		Compatible:            len(diff.BreakingChanges) == 0,
	}

	return diff
}

func getOperations(pathItem *openapi3.PathItem) map[string]*openapi3.Operation {
	ops := make(map[string]*openapi3.Operation)
	if pathItem.Get != nil {
		ops["GET"] = pathItem.Get
	}
	if pathItem.Post != nil {
		ops["POST"] = pathItem.Post
	}
	if pathItem.Put != nil {
		ops["PUT"] = pathItem.Put
	}
	if pathItem.Delete != nil {
		ops["DELETE"] = pathItem.Delete
	}
	if pathItem.Patch != nil {
		ops["PATCH"] = pathItem.Patch
	}
	return ops
}

func comparePathItems(path string, oldItem, newItem *openapi3.PathItem, diff *SpecDiff) {
	oldOps := getOperations(oldItem)
	newOps := getOperations(newItem)

	// Check for removed methods (breaking change)
	for method := range oldOps {
		if _, exists := newOps[method]; !exists {
			diff.BreakingChanges = append(diff.BreakingChanges, Change{
				Type:        "removed_method",
				Path:        path,
				Method:      method,
				Description: fmt.Sprintf("Method %s removed from %s", method, path),
				Severity:    "high",
			})
		}
	}

	// Check for new methods
	for method := range newOps {
		if _, exists := oldOps[method]; !exists {
			diff.ModifiedEndpoints = append(diff.ModifiedEndpoints, Change{
				Type:        "new_method",
				Path:        path,
				Method:      method,
				Description: fmt.Sprintf("New method %s added to %s", method, path),
				Severity:    "info",
			})
		}
	}

	// Compare common methods for changes
	for method, oldOp := range oldOps {
		if newOp, exists := newOps[method]; exists {
			compareOperations(path, method, oldOp, newOp, diff)
		}
	}
}

func compareOperations(path, method string, oldOp, newOp *openapi3.Operation, diff *SpecDiff) {
	// Check for deprecation
	if !oldOp.Deprecated && newOp.Deprecated {
		diff.DeprecatedItems = append(diff.DeprecatedItems, Change{
			Type:        "deprecated",
			Path:        path,
			Method:      method,
			Description: fmt.Sprintf("Endpoint %s %s is now deprecated", method, path),
			Severity:    "medium",
		})
	}

	// Check for required parameter changes
	oldRequired := countRequiredParams(oldOp.Parameters)
	newRequired := countRequiredParams(newOp.Parameters)

	if newRequired > oldRequired {
		diff.BreakingChanges = append(diff.BreakingChanges, Change{
			Type:        "new_required_param",
			Path:        path,
			Method:      method,
			Description: fmt.Sprintf("New required parameters added to %s %s", method, path),
			Severity:    "high",
			OldValue:    fmt.Sprintf("%d required params", oldRequired),
			NewValue:    fmt.Sprintf("%d required params", newRequired),
		})
	}
}

func countRequiredParams(params openapi3.Parameters) int {
	count := 0
	for _, param := range params {
		if param.Value != nil && param.Value.Required {
			count++
		}
	}
	return count
}

func printHumanReadableDiff(diff *SpecDiff) {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║      API Specification Diff Report      ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// Summary
	fmt.Println("📊 Summary:")
	fmt.Printf("  Breaking Changes: %d\n", diff.Summary.TotalBreakingChanges)
	fmt.Printf("  New Endpoints: %d\n", diff.Summary.TotalNewEndpoints)
	fmt.Printf("  Removed Endpoints: %d\n", diff.Summary.TotalRemovedEndpoints)
	fmt.Printf("  Modified Endpoints: %d\n", diff.Summary.TotalModifications)

	if diff.Summary.Compatible {
		fmt.Println("  ✅ Backward Compatible: YES")
	} else {
		fmt.Println("  ❌ Backward Compatible: NO")
	}
	fmt.Println()

	// Breaking changes
	if len(diff.BreakingChanges) > 0 {
		fmt.Println("🚨 Breaking Changes:")
		for _, change := range diff.BreakingChanges {
			fmt.Printf("  ❌ [%s] %s %s\n", change.Type, change.Method, change.Path)
			fmt.Printf("     %s\n", change.Description)
		}
		fmt.Println()
	}

	// New endpoints
	if len(diff.NewEndpoints) > 0 {
		fmt.Println("✨ New Endpoints:")
		for _, change := range diff.NewEndpoints {
			fmt.Printf("  ✅ %s %s\n", change.Method, change.Path)
		}
		fmt.Println()
	}

	// Deprecated items
	if len(diff.DeprecatedItems) > 0 {
		fmt.Println("⚠️  Deprecated:")
		for _, change := range diff.DeprecatedItems {
			fmt.Printf("  ⚠️  %s %s\n", change.Method, change.Path)
		}
		fmt.Println()
	}
}
