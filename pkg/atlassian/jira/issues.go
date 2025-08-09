package jira

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// FetchIssueContext retrieves a single issue by key, returning a concise text block including
// key, summary, status, labels, description, and optionally last N comments.
func (c *client) FetchIssueContext(ctx context.Context, key string, includeComments bool, commentsLimit int) (string, error) {
	if key == "" {
		return "", fmt.Errorf("jira: issue key is required")
	}

	// Expand description and comments to enrich the context
	expands := []string{"renderedFields", "names"}
	issue, _, err := c.Client.Issue.Get(ctx, key, expands, nil)
	if err != nil {
		return "", err
	}

	// Build context
	b := &strings.Builder{}
	fmt.Fprintf(b, "Issue: %s\n", issue.Key)
	if issue.Fields != nil {
		if issue.Fields.Summary != "" {
			fmt.Fprintf(b, "Summary: %s\n", issue.Fields.Summary)
		}
		if issue.Fields.Status != nil && issue.Fields.Status.Name != "" {
			fmt.Fprintf(b, "Status: %s\n", issue.Fields.Status.Name)
		}
		if len(issue.Fields.Labels) > 0 {
			fmt.Fprintf(b, "Labels: %s\n", strings.Join(issue.Fields.Labels, ", "))
		}
		if issue.Fields.Description != "" {
			fmt.Fprintf(b, "Description:\n%s\n", strings.TrimSpace(issue.Fields.Description))
		}
		if includeComments && issue.Fields.Comment != nil && len(issue.Fields.Comment.Comments) > 0 {
			fmt.Fprintf(b, "Comments:\n")
			limit := len(issue.Fields.Comment.Comments)
			if commentsLimit > 0 && commentsLimit < limit {
				limit = commentsLimit
			}
			for i := 0; i < limit; i++ {
				cmt := issue.Fields.Comment.Comments[len(issue.Fields.Comment.Comments)-1-i]
				author := ""
				if cmt.Author != nil {
					author = cmt.Author.DisplayName
				}
				fmt.Fprintf(b, "- %s: %s\n", author, strings.TrimSpace(cmt.Body))
			}
		}
	}

	return b.String(), nil
}

// SearchIssuesContext runs a JQL and returns concatenated contexts for the first N issues.
func (c *client) SearchIssuesContext(ctx context.Context, jql string, maxResults int, includeComments bool) (string, error) {
	if strings.TrimSpace(jql) == "" {
		return "", fmt.Errorf("jira: jql is required")
	}
	if maxResults <= 0 {
		maxResults = 10
	}

	fields := []string{"summary", "status", "labels", "description"}
	// Use the Issue Search service (RichText/v2)
	issuesPage, _, err := c.Client.Issue.Search.Post(ctx, jql, fields, nil, 0, maxResults, "strict")
	if err != nil {
		return "", err
	}

	var blocks []string
	for _, iss := range issuesPage.Issues {
		// Each iss is models.IssueScheme (v3). We need a second call to include comments if needed
		key := iss.Key
		block, err := c.FetchIssueContext(ctx, key, includeComments, 5)
		if err != nil {
			// Continue on individual issue failures
			slog.Warn("Failed to fetch issue", "key", key, "error", err)
			continue
		}
		blocks = append(blocks, block)
	}
	return strings.Join(blocks, "\n\n"), nil
}
