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

// CreatePageInput defines the input parameters for creating a Confluence page
type CreatePageInput struct {
	SpaceKey string `json:"space_key" validate:"required"`
	Title    string `json:"title" validate:"required"`
	Content  string `json:"content" validate:"required"`
	ParentID string `json:"parent_id,omitempty"`
}

// CreatePageOutput defines the output structure for page creation results
type CreatePageOutput struct {
	Success       bool   `json:"success"`
	Title         string `json:"title"`
	ID            string `json:"id"`
	Version       int    `json:"version"`
	Link          string `json:"link"`
	Message       string `json:"message"`
}

// confluenceCreatePageHandler handles the creation of new Confluence pages using typed input
func confluenceCreatePageHandler(ctx context.Context, req mcp.CallToolRequest, input CreatePageInput) (*mcp.CallToolResult, error) {
	client, err := services.ConfluenceClient()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to initialize Confluence client: %v", err)), nil
	}

	// Create page payload
	payload := &models.ContentScheme{
		Type:  "page",
		Title: input.Title,
		Space: &models.SpaceScheme{
			Key: input.SpaceKey,
		},
		Body: &models.BodyScheme{
			Storage: &models.BodyNodeScheme{
				Value:          input.Content,
				Representation: "storage",
			},
		},
	}

	// Handle optional parent ID
	if input.ParentID != "" {
		payload.Ancestors = []*models.ContentScheme{
			{
				ID: input.ParentID,
			},
		}
	}

	// Create the page
	newPage, response, err := client.Content.Create(ctx, payload)
	if err != nil {
		if response != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create page: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to create page: %v", err)), nil
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

	output := CreatePageOutput{
		Success: true,
		Title:   newPage.Title,
		ID:      newPage.ID,
		Version: versionNumber,
		Link:    selfLink,
		Message: fmt.Sprintf("Page created successfully!\nTitle: %s\nID: %s\nVersion: %d\nLink: %s",
			newPage.Title,
			newPage.ID,
			versionNumber,
			selfLink,
		),
	}

	jsonData, err := yaml.Marshal(output)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func RegisterCreatePageTool(s *server.MCPServer) {
	createPageTool := mcp.NewTool("create_page",
		mcp.WithDescription("Create a new Confluence page"),
		mcp.WithString("space_key", mcp.Required(), mcp.Description("The key of the space where the page will be created")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title of the page")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content of the page in storage format (XHTML)")),
		mcp.WithString("parent_id", mcp.Description("ID of the parent page (optional)")),
	)
	s.AddTool(createPageTool, mcp.NewTypedToolHandler(confluenceCreatePageHandler))
} 