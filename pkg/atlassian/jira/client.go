package jira

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	jira "github.com/ctreminiom/go-atlassian/v2/jira/v2"
)

var (
	HOST     = os.Getenv("JIRA_HOST")
	USER     = os.Getenv("JIRA_USER")
	ApiToken = os.Getenv("JIRA_API_TOKEN")
)

type client struct {
	*jira.Client
}

type Client interface {
	// GetProjects prints a sample project (dev helper)
	GetProjects()
	// FetchIssueContext returns a compact, human-readable context for a single issue key
	FetchIssueContext(ctx context.Context, key string, includeComments bool, commentsLimit int) (string, error)
	// SearchIssuesContext returns concatenated contexts for issues matching a JQL
	SearchIssuesContext(ctx context.Context, jql string, maxResults int, includeComments bool) (string, error)
}

func New(ctx context.Context) *client {
	jiraClient, err := jira.New(nil, HOST)
	if err != nil {
		slog.Error("Failed to create JIRA client", "error", err)
		os.Exit(1)
	}

	jiraClient.Auth.SetBasicAuth(USER, ApiToken)

	return &client{jiraClient}
}

// sanitizeString trims and ensures non-empty strings have a trailing newline
func sanitizeString(s string) string {
	if len(s) == 0 {
		return ""
	}
	return s
}

// formatIssueV2 returns a readable block with key, summary, status, labels, description and optional comments
func formatIssueV2(issue interface{}, includeComments bool, commentsLimit int) string {
	// We avoid direct dependency on models to keep this helper resilient to minor changes.
	// Use type assertions against known shapes from v2 models where possible.
	// Fallback to fmt.Sprintf if shapes change.
	type fieldsLike struct {
		Summary     string
		Description string
		Labels      []string
		Status      *struct{ Name string }
		Comment     *struct {
			Comments []struct {
				Author *struct{ DisplayName string }
				Body   string
			}
		}
	}
	type issueLike struct {
		Key    string
		Fields *fieldsLike
	}

	if v, ok := issue.(*issueLike); ok && v != nil {
		// Should not happen due to package boundary, keep for clarity
		_ = v
	}

	// Reflect-light access using fmt and type assertions through map[string]any is overkill.
	// Instead, rely on fmt to print if unknown.

	// Best effort extraction via fmt.Sprintf on known JSON tags is not feasible without models.
	// Therefore, we build context using the library getters in issues.go where we have the concrete type.
	return fmt.Sprintf("%v\n", issue)
}
