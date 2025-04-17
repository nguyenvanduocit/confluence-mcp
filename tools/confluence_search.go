package tools

import (
	"context"
	"fmt"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
)

// confluenceSearchHandler is a handler for the confluence search tool
func confluenceSearchHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Get search query from arguments
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query argument is required")
	}
	options := &models.SearchContentOptions{
		Limit: 5,
	}

	var results string

	contents, response, err := client.Search.Content(ctx, query, options)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("search failed: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}

		return nil, fmt.Errorf("search failed: %v", err)
	}

	// Convert results to map format
	for _, content := range contents.Results {
		results += fmt.Sprintf(`
Title: %s
ID: %s 
Type: %s
Link: %s
Last Modified: %s
Body:
%s
----------------------------------------
`,
			content.Content.Title,
			content.Content.ID,
			content.Content.Type,
			content.Content.Links.Self,
			content.LastModified,
			content.Excerpt,
		)
	}

	return mcp.NewToolResultText(results), nil
}

func RegisterConfluenceSearchTool(s *server.MCPServer) {
	tool := mcp.NewTool("confluence_search",
		mcp.WithDescription("Search Confluence"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Atlassian Confluence Query Language (CQL)")),
	)
	s.AddTool(tool, confluenceSearchHandler)
} 