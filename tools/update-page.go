package tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
	"gopkg.in/yaml.v3"
)

// UpdatePageInput defines the input parameters for updating a Confluence page
type UpdatePageInput struct {
	PageID        string `json:"page_id" validate:"required"`
	Title         string `json:"title,omitempty"`
	Content       string `json:"content,omitempty"`
	VersionNumber string `json:"version_number,omitempty"`
}

// UpdatePageOutput defines the output structure for page update results
type UpdatePageOutput struct {
	Success bool   `json:"success"`
	Title   string `json:"title"`
	ID      string `json:"id"`
	Version int    `json:"version"`
	Link    string `json:"link"`
	Message string `json:"message"`
}

// confluenceUpdatePageHandler handles updating existing Confluence pages
func confluenceUpdatePageHandler(ctx context.Context, request mcp.CallToolRequest, input UpdatePageInput) (*mcp.CallToolResult, error) {
	client, err := services.ConfluenceClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to initialize Confluence client: %v", err)), nil
	}

	// Get the latest version of the page
	currentPage, response, err := client.Content.Get(ctx, input.PageID, []string{"version"}, 0)
	if err != nil {
		if response != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get current page: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to get current page: %v", err)), nil
	}

	// Create update payload
	payload := &models.ContentScheme{
		ID:    input.PageID,
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
	if input.Title != "" {
		payload.Title = input.Title
	}

	// Handle content update
	if input.Content != "" {
		payload.Body = &models.BodyScheme{
			Storage: &models.BodyNodeScheme{
				Value:          input.Content,
				Representation: "storage",
			},
		}
	}

	// Handle version number override
	if input.VersionNumber != "" {
		version, err := strconv.Atoi(input.VersionNumber)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid version_number: %v", err)), nil
		}
		payload.Version.Number = version
	}

	// Update the page
	updatedPage, response, err := client.Content.Update(ctx, input.PageID, payload)
	if err != nil {
		if response != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to update page: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to update page: %v", err)), nil
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

	output := UpdatePageOutput{
		Success: true,
		Title:   updatedPage.Title,
		ID:      updatedPage.ID,
		Version: versionNumber,
		Link:    selfLink,
		Message: fmt.Sprintf("Page updated successfully!\nTitle: %s\nID: %s\nVersion: %d\nLink: %s",
			updatedPage.Title,
			updatedPage.ID,
			versionNumber,
			selfLink,
		),
	}

	// Marshal to YAML
	responseText, err := yaml.Marshal(output)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseText)), nil
}

func RegisterUpdatePageTool(s *server.MCPServer) {
	updatePageTool := mcp.NewTool("update_page",
		mcp.WithDescription("Update an existing Confluence page"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("ID of the page to update")),
		mcp.WithString("title", mcp.Description("New title of the page (optional)")),
		mcp.WithString("content", mcp.Description("New content of the page in storage format (XHTML)")),
		mcp.WithString("version_number", mcp.Description("Version number for optimistic locking (optional)")),
	)
	s.AddTool(updatePageTool, mcp.NewTypedToolHandler(confluenceUpdatePageHandler))
} 