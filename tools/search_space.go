package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
)

// confluenceSearchSpaceHandler searches spaces using CQL
func confluenceSearchSpaceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query argument is required")
	}

	cql := fmt.Sprintf("type=space AND text~\"%s\"", strings.TrimSpace(query))
	options := &models.SearchContentOptions{Limit: 5}

	results, response, err := client.Search.Content(ctx, cql, options)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("search failed: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("search failed: %v", err)
	}

	var out strings.Builder
	for _, r := range results.Results {
		if r.Space != nil {
			link := r.URL
			if link == "" && r.Space.Links != nil {
				link = r.Space.Links.Self
			}
			fmt.Fprintf(&out, "Name: %s\nKey: %s\nLink: %s\n----------------------------------------\n", r.Space.Name, r.Space.Key, link)
		}
	}

	return mcp.NewToolResultText(out.String()), nil
}

// RegisterSearchSpaceTool registers the search_space tool
func RegisterSearchSpaceTool(s *server.MCPServer) {
	tool := mcp.NewTool("search_space",
		mcp.WithDescription("Search Confluence spaces by name using CQL"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Text to search for")),
	)
	s.AddTool(tool, confluenceSearchSpaceHandler)
}
