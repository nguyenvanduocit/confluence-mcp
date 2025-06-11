package tools

import (
	"context"
	"fmt"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
)

func confluenceGetPageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Get page ID from arguments
	pageID, err := request.RequireString("page_id")
	if err != nil {
		return nil, err
	}

	// Request content with various expanded views
	expandParams := []string{"body.storage", "body.view", "body"}
	content, response, err := client.Content.Get(ctx, pageID, expandParams, 0)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to get page: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("failed to get page: %v", err)
	}

	// Process content - use original HTML content
	var htmlContent string
	if content.Body == nil {
		htmlContent = "Page body is nil - no content available"
	} else if content.Body.Storage != nil && content.Body.Storage.Value != "" {
		htmlContent = content.Body.Storage.Value
	} else if content.Body.View != nil && content.Body.View.Value != "" {
		htmlContent = content.Body.View.Value
	} else {
		htmlContent = "No content available in either storage or view format"
	}

	var versionNumber int
	if content.Version != nil {
		versionNumber = content.Version.Number
	}

	result := fmt.Sprintf(`
Title: %s
ID: %s
Version: %d
Type: %s
Content:
%s
`,
		content.Title,
		content.ID,
		versionNumber,
		content.Type,
		htmlContent,
	)

	// Get direct child pages
	childPages, childResponse, err := client.Content.ChildrenDescendant.ChildrenByType(
		ctx,
		pageID,
		"page",
		0,
		[]string{"title", "id", "version"},
		0,
		100,
	)
	
	if err == nil && childResponse != nil && childPages != nil && len(childPages.Results) > 0 {
		result += "\nDirect Child Pages:\n"
		for _, childPage := range childPages.Results {
			var childVersion int
			if childPage.Version != nil {
				childVersion = childPage.Version.Number
			}
			
			result += fmt.Sprintf("- %s (ID: %s, Version: %d)\n", 
				childPage.Title, 
				childPage.ID, 
				childVersion,
			)
		}
	} else if err != nil {
		result += fmt.Sprintf("\nError retrieving child pages: %v\n", err)
	}
	
	// Get all descendants (pages at all levels)
	descendants, descendantsResponse, err := client.Content.ChildrenDescendant.DescendantsByType(
		ctx,
		pageID,
		"page",
		"all",
		[]string{"title", "id", "version"},
		0,
		100,
	)
	
	if err == nil && descendantsResponse != nil && descendants != nil && len(descendants.Results) > 0 {
		// Create a map to track direct children to avoid duplication
		directChildren := make(map[string]bool)
		if childPages != nil && len(childPages.Results) > 0 {
			for _, child := range childPages.Results {
				directChildren[child.ID] = true
			}
		}
		// Add only descendants that aren't direct children
		var nonDirectDescendants []*models.ContentScheme
		for _, descendant := range descendants.Results {
			if !directChildren[descendant.ID] {
				nonDirectDescendants = append(nonDirectDescendants, descendant)
			}
		}
		if len(nonDirectDescendants) > 0 {
			result += "\nAll Descendant Pages (including all levels):\n"
			for _, descendant := range nonDirectDescendants {
				var descendantVersion int
				if descendant.Version != nil {
					descendantVersion = descendant.Version.Number
				}
				result += fmt.Sprintf("- %s (ID: %s, Version: %d)\n",
					descendant.Title,
					descendant.ID,
					descendantVersion,
				)
			}
		}
	} else if err != nil {
		result += fmt.Sprintf("\nError retrieving descendant pages: %v\n", err)
	}

	return mcp.NewToolResultText(result), nil
}

func RegisterGetPageTool(s *server.MCPServer) {
	pageTool := mcp.NewTool("get_page",
		mcp.WithDescription("Get Confluence page content"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID")),
	)
	s.AddTool(pageTool, confluenceGetPageHandler)
} 