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

// confluenceListSpacesHandler lists spaces with optional filters
func confluenceListSpacesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	startAt := 0
	if v, ok := request.Params.Arguments["start_at"].(float64); ok {
		startAt = int(v)
	}
	maxResults := 25
	if v, ok := request.Params.Arguments["max_results"].(float64); ok {
		maxResults = int(v)
	}

	options := &models.GetSpacesOptionScheme{}
	if status, ok := request.Params.Arguments["status"].(string); ok && status != "" {
		options.Status = status
	}
	if spaceType, ok := request.Params.Arguments["space_type"].(string); ok && spaceType != "" {
		options.SpaceType = spaceType
	}
	if expand, ok := request.Params.Arguments["expand"].(string); ok && expand != "" {
		options.Expand = strings.Split(expand, ",")
	}

	spaces, response, err := client.Space.Gets(ctx, options, startAt, maxResults)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to list spaces: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("failed to list spaces: %v", err)
	}

	var out strings.Builder
	for _, sp := range spaces.Results {
		link := ""
		if sp.Links != nil {
			link = sp.Links.Self
		}
		fmt.Fprintf(&out, "Name: %s\nKey: %s\nID: %d\nLink: %s\n----------------------------------------\n", sp.Name, sp.Key, sp.ID, link)
	}

	return mcp.NewToolResultText(out.String()), nil
}

// RegisterListSpacesTool registers the list_spaces tool
func RegisterListSpacesTool(s *server.MCPServer) {
	tool := mcp.NewTool("list_spaces",
		mcp.WithDescription("List Confluence spaces"),
		mcp.WithNumber("start_at", mcp.Description("Pagination start")),
		mcp.WithNumber("max_results", mcp.Description("Max results, default 25")),
		mcp.WithString("status", mcp.Description("Space status filter")),
		mcp.WithString("space_type", mcp.Description("Space type filter")),
		mcp.WithString("expand", mcp.Description("Fields to expand")),
	)
	s.AddTool(tool, confluenceListSpacesHandler)
}
