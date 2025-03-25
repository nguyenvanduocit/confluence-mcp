package tools

import (
	"context"
	"fmt"
	"strconv"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/nguyenvanduocit/confluence-mcp/services"
	"github.com/nguyenvanduocit/confluence-mcp/util"
)

// registerConfluenceTool is a function that registers the confluence tools to the server
func RegisterConfluenceTool(s *server.MCPServer) {
	tool := mcp.NewTool("confluence_search",
		mcp.WithDescription("Search Confluence"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Atlassian Confluence Query Language (CQL)")),
	)

	s.AddTool(tool, confluenceSearchHandler)

	// Add new tool for getting page content
	pageTool := mcp.NewTool("confluence_get_page",
		mcp.WithDescription("Get Confluence page content"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID")),
	)
	s.AddTool(pageTool, util.ErrorGuard(confluencePageHandler))

	// Add new tool for creating Confluence pages
	createPageTool := mcp.NewTool("confluence_create_page",
		mcp.WithDescription("Create a new Confluence page"),
		mcp.WithString("space_key", mcp.Required(), mcp.Description("The key of the space where the page will be created")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title of the page")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content of the page in storage format (XHTML)")),
		mcp.WithString("parent_id", mcp.Description("ID of the parent page (optional)")),
	)
	s.AddTool(createPageTool, util.ErrorGuard(confluenceCreatePageHandler))

	// Add new tool for updating Confluence pages
	updatePageTool := mcp.NewTool("confluence_update_page",
		mcp.WithDescription("Update an existing Confluence page"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("ID of the page to update")),
		mcp.WithString("title", mcp.Description("New title of the page (optional)")),
		mcp.WithString("content", mcp.Description("New content of the page in storage format (XHTML)")),
		mcp.WithString("version_number", mcp.Description("Version number for optimistic locking (optional)")),
	)
	s.AddTool(updatePageTool, util.ErrorGuard(confluenceUpdatePageHandler))
}

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

func confluencePageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Get page ID from arguments
	pageID, ok := request.Params.Arguments["page_id"].(string)
	if !ok {
		return nil, fmt.Errorf("page_id argument is required")
	}

	content, response, err := client.Content.Get(ctx, pageID, []string{"body.storage"}, 1)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to get page: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("failed to get page: %v", err)
	}

	mdContent, err := htmltomarkdown.ConvertString(content.Body.Storage.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HTML to Markdown: %v", err)
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
		content.Version.Number,
		content.Type,
		mdContent,
	)

	return mcp.NewToolResultText(result), nil
}

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

	result := fmt.Sprintf("Page created successfully!\nTitle: %s\nID: %s\nVersion: %d\nLink: %s",
		newPage.Title,
		newPage.ID,
		newPage.Version.Number,
		newPage.Links.Self,
	)

	return mcp.NewToolResultText(result), nil
}

// confluenceUpdatePageHandler handles updating existing Confluence pages
func confluenceUpdatePageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Extract required arguments
	pageID, ok := request.Params.Arguments["page_id"].(string)
	if !ok {
		return nil, fmt.Errorf("page_id argument is required")
	}

	currentPage, response, err := client.Content.Get(ctx, pageID, []string{"version"}, 1)
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
		Version: &models.ContentVersionScheme{
			Number: currentPage.Version.Number + 1,
		},
	}

	// Handle optional title update
	if title, ok := request.Params.Arguments["title"].(string); ok && title != "" {
		payload.Title = title
	}

	// Handle content update
	if content, ok := request.Params.Arguments["content"].(string); ok && content != "" {
		payload.Body = &models.BodyScheme{
			Storage: &models.BodyNodeScheme{
				Value:          content,
				Representation: "storage",
			},
		}
	}

	// Handle version number override
	if versionStr, ok := request.Params.Arguments["version_number"].(string); ok && versionStr != "" {
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

	result := fmt.Sprintf("Page updated successfully!\nTitle: %s\nID: %s\nVersion: %d\nLink: %s",
		updatedPage.Title,
		updatedPage.ID,
		updatedPage.Version.Number,
		updatedPage.Links.Self,
	)

	return mcp.NewToolResultText(result), nil
}
