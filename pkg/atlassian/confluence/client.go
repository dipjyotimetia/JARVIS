package confluence

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"

	confluenceClassic "github.com/ctreminiom/go-atlassian/v2/confluence"
	confluence "github.com/ctreminiom/go-atlassian/v2/confluence/v2"
	models "github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
)

var (
	HOST     = os.Getenv("CONFLUENCE_HOST")
	USER     = os.Getenv("CONFLUENCE_USER")
	ApiToken = os.Getenv("CONFLUENCE_API_TOKEN")
)

type client struct {
	*confluence.Client
}

type Client interface {
	GetPages() ([]string, error)
	// FetchPageContext returns a concise, human-readable context for a Confluence page.
	// bodyFormat can be "storage" or "atlas_doc_format". If empty, defaults to "storage".
	// maxChars trims the body to a maximum length (0 means no trim).
	FetchPageContext(ctx context.Context, pageID int, bodyFormat string, maxChars int) (string, error)
	// SearchPagesContext runs a CQL query and returns concatenated contexts for the first N results.
	// Filters to type=page. bodyFormat as in FetchPageContext. maxChars trims each page body.
	SearchPagesContext(ctx context.Context, cql string, maxResults int, bodyFormat string, maxChars int) (string, error)
}

func New(ctx context.Context) *client {
	confluenceClient, err := confluence.New(nil, HOST)
	if err != nil {
		slog.Error("Failed to create Confluence client", "error", err)
		os.Exit(1)
	}

	confluenceClient.Auth.SetBasicAuth(USER, ApiToken)

	return &client{confluenceClient}
}

func (c *client) GetPages() ([]string, error) {
	pages, _, err := c.Client.Page.Gets(context.Background(), nil, "", 0)
	if err != nil {
		return nil, err
	}

	var pageTitles []string
	for _, page := range pages.Results {
		pageTitles = append(pageTitles, page.Title)
	}

	return pageTitles, nil
}

// FetchPageContext retrieves a single page and returns a concise text block:
// Title and cleaned body content (tags stripped, whitespace normalized).
func (c *client) FetchPageContext(ctx context.Context, pageID int, bodyFormat string, maxChars int) (string, error) {
	if pageID <= 0 {
		return "", fmt.Errorf("confluence: pageID is required")
	}
	if strings.TrimSpace(bodyFormat) == "" {
		bodyFormat = "storage"
	}

	page, _, err := c.Client.Page.Get(ctx, pageID, bodyFormat, false, 0)
	if err != nil {
		return "", err
	}

	var title string
	var raw string
	if page != nil {
		title = strings.TrimSpace(page.Title)
		if page.Body != nil {
			switch strings.ToLower(bodyFormat) {
			case "storage":
				if page.Body.Storage != nil {
					raw = page.Body.Storage.Value
				}
			case "atlas_doc_format", "adf":
				if page.Body.AtlasDocFormat != nil {
					raw = page.Body.AtlasDocFormat.Value
				}
			default:
				if page.Body.Storage != nil {
					raw = page.Body.Storage.Value
				}
			}
		}
	}

	cleaned := normalizeText(stripTags(raw))
	if maxChars > 0 && len(cleaned) > maxChars {
		cleaned = cleaned[:maxChars] + "…"
	}

	b := &strings.Builder{}
	if title != "" {
		fmt.Fprintf(b, "Page: %s\n", title)
	} else {
		fmt.Fprintf(b, "Page ID: %d\n", pageID)
	}
	if cleaned != "" {
		fmt.Fprintf(b, "Content:\n%s\n", cleaned)
	}
	return b.String(), nil
}

// SearchPagesContext runs a CQL search, filters results to pages, and concatenates
// page contexts using FetchPageContext.
func (c *client) SearchPagesContext(ctx context.Context, cql string, maxResults int, bodyFormat string, maxChars int) (string, error) {
	if strings.TrimSpace(cql) == "" {
		return "", fmt.Errorf("confluence: cql is required")
	}
	if maxResults <= 0 {
		maxResults = 10
	}

	// Use classic REST search to get content by CQL, then hydrate each result via v2 Page.Get
	classicClient, err := confluenceClassic.New(nil, HOST)
	if err != nil {
		return "", err
	}
	classicClient.Auth.SetBasicAuth(USER, ApiToken)

	pageSet, _, err := classicClient.Search.Content(ctx, cql, &models.SearchContentOptions{Limit: maxResults})
	if err != nil {
		return "", err
	}

	var blocks []string
	if pageSet != nil {
		for _, item := range pageSet.Results {
			if item == nil || item.Content == nil || strings.ToLower(strings.TrimSpace(item.Content.Type)) != "page" {
				continue
			}
			// item.Content.ID is a string
			idStr := strings.TrimSpace(item.Content.ID)
			if idStr == "" {
				continue
			}
			idInt, convErr := strconv.Atoi(idStr)
			if convErr != nil || idInt == 0 {
				continue
			}
			block, ferr := c.FetchPageContext(ctx, idInt, bodyFormat, maxChars)
			if ferr != nil {
				continue
			}
			if strings.TrimSpace(block) != "" {
				blocks = append(blocks, block)
			}
		}
	}

	return strings.Join(blocks, "\n\n"), nil
}

var (
	tagStripper   = regexp.MustCompile(`(?s)<[^>]*>`) // crude HTML/XML tag remover
	spaceCondense = regexp.MustCompile(`[\t\x0B\f\r ]+`)
)

func stripTags(s string) string {
	if s == "" {
		return s
	}
	return tagStripper.ReplaceAllString(s, " ")
}

func normalizeText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// Collapse whitespace and normalize newlines
	s = spaceCondense.ReplaceAllString(s, " ")
	// Unescape common HTML entities minimally if present (leave advanced cases as-is)
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	return strings.TrimSpace(s)
}
