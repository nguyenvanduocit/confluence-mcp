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

// GetPageInput defines the input parameters for getting a Confluence page
type GetPageInput struct {
	PageID string `json:"page_id" validate:"required"`
}

// GetPageOutput defines the output structure for page retrieval results
type GetPageOutput struct {
	Title           string      `json:"title"`
	ID              string      `json:"id"`
	Version         int         `json:"version"`
	Type            string      `json:"type"`
	Content         string      `json:"content"`
	DirectChildren  []PageInfo  `json:"direct_children,omitempty"`
	AllDescendants  []PageInfo  `json:"all_descendants,omitempty"`
	Message         string      `json:"message"`
}

// PageInfo represents basic page information
type PageInfo struct {
	Title   string `json:"title"`
	ID      string `json:"id"`
	Version int    `json:"version"`
}

func confluenceGetPageHandler(ctx context.Context, request mcp.CallToolRequest, input GetPageInput) (*mcp.CallToolResult, error) {
	client, err := services.ConfluenceClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to initialize Confluence client: %v", err)), nil
	}

	// Request content with various expanded views
	expandParams := []string{"body.storage", "body.view", "body"}
	content, response, err := client.Content.Get(ctx, input.PageID, expandParams, 0)
	if err != nil {
		if response != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get page: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to get page: %v", err)), nil
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

	output := GetPageOutput{
		Title:   content.Title,
		ID:      content.ID,
		Version: versionNumber,
		Type:    content.Type,
		Content: htmlContent,
	}

	// Get direct child pages
	childPages, childResponse, err := client.Content.ChildrenDescendant.ChildrenByType(
		ctx,
		input.PageID,
		"page",
		0,
		[]string{"title", "id", "version"},
		0,
		100,
	)
	
	if err == nil && childResponse != nil && childPages != nil && len(childPages.Results) > 0 {
		output.DirectChildren = make([]PageInfo, 0, len(childPages.Results))
		for _, childPage := range childPages.Results {
			var childVersion int
			if childPage.Version != nil {
				childVersion = childPage.Version.Number
			}
			
			output.DirectChildren = append(output.DirectChildren, PageInfo{
				Title:   childPage.Title,
				ID:      childPage.ID,
				Version: childVersion,
			})
		}
	}
	
	// Get all descendants (pages at all levels)
	descendants, descendantsResponse, err := client.Content.ChildrenDescendant.DescendantsByType(
		ctx,
		input.PageID,
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
			output.AllDescendants = make([]PageInfo, 0, len(nonDirectDescendants))
			for _, descendant := range nonDirectDescendants {
				var descendantVersion int
				if descendant.Version != nil {
					descendantVersion = descendant.Version.Number
				}
				output.AllDescendants = append(output.AllDescendants, PageInfo{
					Title:   descendant.Title,
					ID:      descendant.ID,
					Version: descendantVersion,
				})
			}
		}
	}

	// Set success message
	childCount := len(output.DirectChildren)
	descendantCount := len(output.AllDescendants)
	output.Message = fmt.Sprintf("Page retrieved successfully with %d direct children and %d other descendants", childCount, descendantCount)

	// Marshal to YAML
	responseText, err := yaml.Marshal(output)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseText)), nil
}

func RegisterGetPageTool(s *server.MCPServer) {
	pageTool := mcp.NewTool("get_page",
		mcp.WithDescription("Get Confluence page content"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID")),
	)
	s.AddTool(pageTool, mcp.NewTypedToolHandler(confluenceGetPageHandler))
} 