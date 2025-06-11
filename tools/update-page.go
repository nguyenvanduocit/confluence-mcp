package tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
)

// confluenceUpdatePageHandler handles updating existing Confluence pages
func confluenceUpdatePageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Extract required arguments
	pageID, err := request.RequireString("page_id")
	if err != nil {
		return nil, err
	}

	// Get the latest version of the page
	currentPage, response, err := client.Content.Get(ctx, pageID, []string{"version"}, 0)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to get current page: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("failed to get current page: %v", err)
	}

	// Create update payload
	payload := &models.ContentScheme{
		ID:    pageID,
		Type:  "page",
		Title: currentPage.Title, // Keep existing title by default
	}
	
	// Safely handle version increment
	if currentPage.Version != nil {
		payload.Version = &models.ContentVersionScheme{
			Number: currentPage.Version.Number + 1,
		}
	} else {
		payload.Version = &models.ContentVersionScheme{
			Number: 1, // Default to version 1 if no version info
		}
	}

	// Handle optional title update
	if title := request.GetString("title", ""); title != "" {
		payload.Title = title
	}

	// Handle content update
	if content := request.GetString("content", ""); content != "" {
		payload.Body = &models.BodyScheme{
			Storage: &models.BodyNodeScheme{
				Value:          content,
				Representation: "storage",
			},
		}
	}

	// Handle version number override
	if versionStr := request.GetString("version_number", ""); versionStr != "" {
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return nil, fmt.Errorf("invalid version_number: %v", err)
		}
		payload.Version.Number = version
	}

	// Update the page
	updatedPage, response, err := client.Content.Update(ctx, pageID, payload)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to update page: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("failed to update page: %v", err)
	}

	var versionNumber int
	if updatedPage.Version != nil {
		versionNumber = updatedPage.Version.Number
	}

	// Check if the Links field is nil
	var selfLink string
	if updatedPage.Links != nil {
		selfLink = updatedPage.Links.Self
	}

	result := fmt.Sprintf("Page updated successfully!\nTitle: %s\nID: %s\nVersion: %d\nLink: %s",
		updatedPage.Title,
		updatedPage.ID,
		versionNumber,
		selfLink,
	)

	return mcp.NewToolResultText(result), nil
}

func RegisterUpdatePageTool(s *server.MCPServer) {
	updatePageTool := mcp.NewTool("update_page",
		mcp.WithDescription("Update an existing Confluence page"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("ID of the page to update")),
		mcp.WithString("title", mcp.Description("New title of the page (optional)")),
		mcp.WithString("content", mcp.Description("New content of the page in storage format (XHTML)")),
		mcp.WithString("version_number", mcp.Description("Version number for optimistic locking (optional)")),
	)
	s.AddTool(updatePageTool, confluenceUpdatePageHandler)
} 