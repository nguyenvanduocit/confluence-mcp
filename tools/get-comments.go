package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
)

// confluenceGetCommentsHandler handles retrieving comments for a Confluence page
func confluenceGetCommentsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Get page ID from arguments
	pageID, err := request.RequireString("page_id")
	if err != nil {
		return nil, err
	}

	// Get optional parameters
	expand := make([]string, 0)
	if expandVal := request.GetString("expand", ""); expandVal != "" {
		expand = append(expand, expandVal)
	}

	// Get optional location parameters
	location := make([]string, 0)
	if locationVal := request.GetString("location", ""); locationVal != "" {
		location = append(location, locationVal)
	}

	// Get optional pagination parameters
	startAt := 0
	if startAtVal := request.GetInt("start_at", 0); startAtVal != 0 {
		startAt = int(startAtVal)
	}

	maxResults := 50
	if maxResultsVal := request.GetInt("max_results", 50); maxResultsVal != 0 {
		maxResults = maxResultsVal
	}

	// Get comments
	comments, response, err := client.Content.Comment.Gets(ctx, pageID, expand, location, startAt, maxResults)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to get comments: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)
		}
		return nil, fmt.Errorf("failed to get comments: %v", err)
	}

	// Format the result
	result := fmt.Sprintf("Comments for page ID %s:\n\n", pageID)

	if len(comments.Results) == 0 {
		result += "No comments found."
	} else {
		for i, comment := range comments.Results {
			result += fmt.Sprintf("Comment #%d:\n", i+1)
			result += fmt.Sprintf("ID: %s\n", comment.ID)
			result += fmt.Sprintf("Title: %s\n", comment.Title)
			result += fmt.Sprintf("Status: %s\n", comment.Status)
			
			// Add author info if available
			if comment.Version != nil && comment.Version.By != nil {
				result += fmt.Sprintf("Author: %s\n", comment.Version.By.DisplayName)
				result += fmt.Sprintf("Created: %s\n", comment.Version.When)
			}
			
			// Add comment body if available
			if comment.Body != nil && comment.Body.View != nil {
				result += fmt.Sprintf("Content: %s\n", comment.Body.View.Value)
			}
			
			result += "----------------------------------------\n"
		}
		
		// Add pagination info
		result += fmt.Sprintf("\nShowing %d of %d comments (page %d).", 
			len(comments.Results), 
			comments.Size,
			(startAt/maxResults)+1)
	}

	return mcp.NewToolResultText(result), nil
}

// RegisterGetCommentsPageTool registers the get_comments tool with the server
func RegisterGetCommentsPageTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_comments",
		mcp.WithDescription("Get comments from a Confluence page"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID")),
		mcp.WithString("expand", mcp.Description("Properties to expand in the response (comma-separated)")),
		mcp.WithString("location", mcp.Description("Comment location filter (inline, footer, resolved)")),
		mcp.WithNumber("start_at", mcp.Description("Starting index for pagination")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of results to return (default: 50)")),
	)
	s.AddTool(tool, confluenceGetCommentsHandler)
} 