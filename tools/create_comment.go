package tools

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
)

// confluenceCreateCommentHandler creates a comment on a page
func confluenceCreateCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := services.ConfluenceClient()

	pageID, ok := request.Params.Arguments["page_id"].(string)
	if !ok {
		return nil, fmt.Errorf("page_id argument is required")
	}
	content, ok := request.Params.Arguments["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content argument is required")
	}

	payload := map[string]interface{}{
		"type": "comment",
		"body": map[string]interface{}{
			"storage": map[string]interface{}{
				"value":          content,
				"representation": "storage",
			},
		},
	}

	endpoint := fmt.Sprintf("wiki/rest/api/content/%s/child/comment", pageID)
	req, err := client.NewRequest(ctx, http.MethodPost, endpoint, "", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %v", err)
	}

	result := new(models.ContentScheme)
	resp, err := client.Call(req, result)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("failed to create comment: %s (endpoint: %s)", resp.Bytes.String(), resp.Endpoint)
		}
		return nil, fmt.Errorf("failed to create comment: %v", err)
	}

	link := ""
	if result.Links != nil {
		link = result.Links.Self
	}

	output := fmt.Sprintf("Comment created successfully!\nID: %s\nLink: %s", result.ID, link)
	return mcp.NewToolResultText(output), nil
}

// RegisterCreateCommentTool registers the create_comment tool
func RegisterCreateCommentTool(s *server.MCPServer) {
	tool := mcp.NewTool("create_comment",
		mcp.WithDescription("Create a comment on a Confluence page"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("ID of the page")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Comment body in storage format")),
	)
	s.AddTool(tool, confluenceCreateCommentHandler)
}
