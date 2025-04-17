package tools

import (
	"context"
	"fmt"
	"strconv"

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
			// Skip if it's a direct child (already displayed above)
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

// confluenceUpdatePageHandler handles updating existing Confluence pages
func confluenceUpdatePageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Extract required arguments
	pageID, ok := request.Params.Arguments["page_id"].(string)
	if !ok {
		return nil, fmt.Errorf("page_id argument is required")
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
