package tools

import (
	"context"
	"fmt"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
)

// confluenceCreatePageHandler handles the creation of new Confluence pages
func confluenceCreatePageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Extract required arguments
	spaceKey, err := request.RequireString("space_key")
	if err != nil {
		return nil, err
	}

	title, err := request.RequireString("title")
	if err != nil {
		return nil, err
	}

	content, err := request.RequireString("content")
	if err != nil {
		return nil, err
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
	if parentID := request.GetString("parent_id", ""); parentID != "" {
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

func RegisterCreatePageTool(s *server.MCPServer) {
	createPageTool := mcp.NewTool("create_page",
		mcp.WithDescription("Create a new Confluence page"),
		mcp.WithString("space_key", mcp.Required(), mcp.Description("The key of the space where the page will be created")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title of the page")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content of the page in storage format (XHTML)")),
		mcp.WithString("parent_id", mcp.Description("ID of the parent page (optional)")),
	)
	s.AddTool(createPageTool, confluenceCreatePageHandler)
} 