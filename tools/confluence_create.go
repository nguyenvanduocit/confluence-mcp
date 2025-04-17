package tools

import (
	"context"
	"fmt"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
	"github.com/nguyenvanduocit/confluence-mcp/util"
)

// confluenceCreatePageHandler handles the creation of new Confluence pages
func confluenceCreatePageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Extract required arguments
	spaceKey, ok := request.Params.Arguments["space_key"].(string)
	if !ok {
		return nil, fmt.Errorf("space_key argument is required")
	}

	title, ok := request.Params.Arguments["title"].(string)
	if !ok {
		return nil, fmt.Errorf("title argument is required")
	}

	content, ok := request.Params.Arguments["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content argument is required")
	}

	// Create page payload
	payload := &models.ContentScheme{
		Type:  "page",
		Title: title,
		Space: &models.SpaceScheme{
			Key: spaceKey,
		},
		Body: &models.BodyScheme{
			Storage: &models.BodyNodeScheme{
				Value:          content,
				Representation: "storage",
			},
		},
	}

	// Handle optional parent ID
	if parentID, ok := request.Params.Arguments["parent_id"].(string); ok && parentID != "" {
		payload.Ancestors = []*models.ContentScheme{
			{
				ID: parentID,
			},
		}
	}

	// Create the page
	newPage, response, err := client.Content.Create(ctx, payload)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to create page: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("failed to create page: %v", err)
	}

	var versionNumber int
	if newPage.Version != nil {
		versionNumber = newPage.Version.Number
	}

	// Check if the Links field is nil
	var selfLink string
	if newPage.Links != nil {
		selfLink = newPage.Links.Self
	}

	result := fmt.Sprintf("Page created successfully!\nTitle: %s\nID: %s\nVersion: %d\nLink: %s",
		newPage.Title,
		newPage.ID,
		versionNumber,
		selfLink,
	)

	return mcp.NewToolResultText(result), nil
}

func RegisterConfluenceCreatePageTool(s *server.MCPServer) {
	createPageTool := mcp.NewTool("confluence_create_page",
		mcp.WithDescription("Create a new Confluence page"),
		mcp.WithString("space_key", mcp.Required(), mcp.Description("The key of the space where the page will be created")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title of the page")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content of the page in storage format (XHTML)")),
		mcp.WithString("parent_id", mcp.Description("ID of the parent page (optional)")),
	)
	s.AddTool(createPageTool, util.ErrorGuard(confluenceCreatePageHandler))
} 