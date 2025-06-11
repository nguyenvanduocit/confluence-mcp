package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
	"gopkg.in/yaml.v3"
)

// GetCommentsInput defines the input parameters for getting comments
type GetCommentsInput struct {
	PageID     string `json:"page_id" validate:"required"`
	Expand     string `json:"expand,omitempty"`
	Location   string `json:"location,omitempty"`
	StartAt    int    `json:"start_at,omitempty"`
	MaxResults int    `json:"max_results,omitempty"`
}

// GetCommentsOutput defines the output structure for comments
type GetCommentsOutput struct {
	PageID      string                   `json:"page_id"`
	Comments    []CommentInfo           `json:"comments"`
	TotalCount  int                     `json:"total_count"`
	CurrentPage int                     `json:"current_page"`
	Message     string                  `json:"message"`
}

// CommentInfo represents a single comment
type CommentInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Author      string `json:"author,omitempty"`
	Created     string `json:"created,omitempty"`
	Content     string `json:"content,omitempty"`
}

// confluenceGetCommentsTypedHandler handles retrieving comments for a Confluence page using typed approach
func confluenceGetCommentsTypedHandler(ctx context.Context, req mcp.CallToolRequest, input GetCommentsInput) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	// Set default values
	if input.MaxResults == 0 {
		input.MaxResults = 50
	}

	// Prepare expand and location parameters
	expand := make([]string, 0)
	if input.Expand != "" {
		expand = append(expand, input.Expand)
	}

	location := make([]string, 0)
	if input.Location != "" {
		location = append(location, input.Location)
	}

	// Get comments
	comments, response, err := client.Content.Comment.Gets(ctx, input.PageID, expand, location, input.StartAt, input.MaxResults)
	if err != nil {
		if response != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get comments: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to get comments: %v", err)), nil
	}

	// Convert to output format
	output := GetCommentsOutput{
		PageID:      input.PageID,
		Comments:    make([]CommentInfo, 0, len(comments.Results)),
		TotalCount:  comments.Size,
		CurrentPage: (input.StartAt / input.MaxResults) + 1,
	}

	if len(comments.Results) == 0 {
		output.Message = "No comments found."
	} else {
		for _, comment := range comments.Results {
			commentInfo := CommentInfo{
				ID:     comment.ID,
				Title:  comment.Title,
				Status: comment.Status,
			}

			// Add author info if available
			if comment.Version != nil && comment.Version.By != nil {
				commentInfo.Author = comment.Version.By.DisplayName
				commentInfo.Created = comment.Version.When
			}

			// Add comment body if available
			if comment.Body != nil && comment.Body.View != nil {
				commentInfo.Content = comment.Body.View.Value
			}

			output.Comments = append(output.Comments, commentInfo)
		}

		output.Message = fmt.Sprintf("Found %d comments (page %d)", len(comments.Results), output.CurrentPage)
	}

	// Marshal to JSON
	responseText, err := yaml.Marshal(output)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseText)), nil
}

// RegisterGetCommentsPageTool registers the get_comments tool with the server using typed handler
func RegisterGetCommentsPageTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_comments",
		mcp.WithDescription("Get comments from a Confluence page"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID")),
		mcp.WithString("expand", mcp.Description("Properties to expand in the response (comma-separated)")),
		mcp.WithString("location", mcp.Description("Comment location filter (inline, footer, resolved)")),
		mcp.WithNumber("start_at", mcp.Description("Starting index for pagination")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of results to return (default: 50)")),
	)
	
	// Use typed tool handler
	s.AddTool(tool, mcp.NewTypedToolHandler(confluenceGetCommentsTypedHandler))
} 