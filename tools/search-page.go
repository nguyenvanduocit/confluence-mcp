package tools

import (
	"context"
	"fmt"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
	"gopkg.in/yaml.v3"
)

// SearchPageInput defines the input parameters for searching Confluence pages
type SearchPageInput struct {
	Query string `json:"query" validate:"required"`
}

// SearchPageOutput defines the output structure for search results
type SearchPageOutput struct {
	Query       string       `json:"query"`
	Results     []SearchResult `json:"results"`
	ResultCount int          `json:"result_count"`
	Message     string       `json:"message"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Title        string `json:"title"`
	ID           string `json:"id"`
	Type         string `json:"type"`
	Link         string `json:"link"`
	LastModified string `json:"last_modified"`
	Excerpt      string `json:"excerpt"`
}

// confluenceSearchHandler is a handler for the confluence search tool
func confluenceSearchHandler(ctx context.Context, request mcp.CallToolRequest, input SearchPageInput) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	options := &models.SearchContentOptions{
		Limit: 5,
	}

	contents, response, err := client.Search.Content(ctx, input.Query, options)
	if err != nil {
		if response != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	output := SearchPageOutput{
		Query:       input.Query,
		Results:     make([]SearchResult, 0, len(contents.Results)),
		ResultCount: len(contents.Results),
	}

	if len(contents.Results) == 0 {
		output.Message = "No results found for the search query"
	} else {
		// Convert results to structured format
		for _, content := range contents.Results {
			result := SearchResult{
				Title:        content.Content.Title,
				ID:           content.Content.ID,
				Type:         content.Content.Type,
				Link:         content.Content.Links.Self,
				LastModified: content.LastModified,
				Excerpt:      content.Excerpt,
			}
			output.Results = append(output.Results, result)
		}
		output.Message = fmt.Sprintf("Found %d results for query: %s", len(contents.Results), input.Query)
	}

	// Marshal to YAML
	responseText, err := yaml.Marshal(output)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseText)), nil
}

func RegisterSearchPageTool(s *server.MCPServer) {
	tool := mcp.NewTool("search_page",
		mcp.WithDescription("Search pages in Confluence"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Atlassian Confluence Query Language (CQL)")),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(confluenceSearchHandler))
} 