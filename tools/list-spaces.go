package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/services"
	"gopkg.in/yaml.v3"
)

// ListSpacesInput defines the input parameters for listing Confluence spaces
type ListSpacesInput struct {
	// No required parameters for listing spaces
}

// ListSpacesOutput defines the output structure for spaces listing results
type ListSpacesOutput struct {
	Spaces      []SpaceInfo `json:"spaces"`
	SpaceCount  int         `json:"space_count"`
	Message     string      `json:"message"`
}

// SpaceInfo represents basic space information
type SpaceInfo struct {
	Key    string `json:"key"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Link   string `json:"link"`
}

// confluenceListSpacesHandler handles listing all Confluence spaces
func confluenceListSpacesHandler(ctx context.Context, request mcp.CallToolRequest, input ListSpacesInput) (*mcp.CallToolResult, error) {
    client, err := services.ConfluenceClient()
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("failed to initialize Confluence client: %v", err)), nil
    }

    // Fetch spaces – default options, first 100 results
    spaces, response, err := client.Space.Gets(ctx, nil, 0, 100)
    if err != nil {
        if response != nil {
            return mcp.NewToolResultError(fmt.Sprintf("failed to list spaces: %s (endpoint: %s)", response.Bytes.String(), response.Endpoint)), nil
        }
        return mcp.NewToolResultError(fmt.Sprintf("failed to list spaces: %v", err)), nil
    }

    output := ListSpacesOutput{
        Spaces:     make([]SpaceInfo, 0, len(spaces.Results)),
        SpaceCount: len(spaces.Results),
    }

    if len(spaces.Results) == 0 {
        output.Message = "No spaces found"
    } else {
        for _, space := range spaces.Results {
            var link string
            if space.Links != nil {
                link = space.Links.Self
            }

            spaceInfo := SpaceInfo{
                Key:    space.Key,
                ID:     space.ID,
                Name:   space.Name,
                Type:   space.Type,
                Status: space.Status,
                Link:   link,
            }
            output.Spaces = append(output.Spaces, spaceInfo)
        }
        output.Message = fmt.Sprintf("Found %d spaces", len(spaces.Results))
    }

    // Marshal to YAML
    responseText, err := yaml.Marshal(output)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
    }

    return mcp.NewToolResultText(string(responseText)), nil
}

// RegisterListSpacesTool registers the list_spaces tool with the MCP server
func RegisterListSpacesTool(s *server.MCPServer) {
    tool := mcp.NewTool("list_spaces",
        mcp.WithDescription("List Confluence spaces"),
    )
    s.AddTool(tool, mcp.NewTypedToolHandler(confluenceListSpacesHandler))
} 