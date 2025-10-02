# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based Model Context Protocol (MCP) server for Atlassian Confluence integration. It provides a standardized interface for AI assistants to interact with Confluence APIs for searching, creating, updating, and managing Confluence pages and spaces.

## Development Commands

### Build
```bash
# Build binary
just build
# Or manually:
CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/confluence-mcp ./main.go
```

### Run Development Server
```bash
# Run with HTTP transport on port 3003
just dev
# Or manually:
go run main.go --env .env --http_port 3003

# Run with stdio transport (for MCP clients)
go run main.go --env .env
```

### Install
```bash
just install
# Or:
go install ./...
```

### Docker
```bash
# Build Docker image
docker build -t confluence-mcp .

# Run with HTTP transport
docker run -p 8080:8080 \
  -e ATLASSIAN_HOST=your_instance.atlassian.net \
  -e ATLASSIAN_EMAIL=your_email@example.com \
  -e ATLASSIAN_TOKEN=your_api_token \
  confluence-mcp --http_port 8080

# Run with stdio transport
docker run --rm -i \
  -e ATLASSIAN_HOST=your_instance.atlassian.net \
  -e ATLASSIAN_EMAIL=your_email@example.com \
  -e ATLASSIAN_TOKEN=your_api_token \
  confluence-mcp
```

## Architecture

### Core Components

1. **Transport Layer** (`main.go`)
   - Supports two transport methods: stdio (default) and HTTP server
   - HTTP server uses mcp-go's StreamableHTTPServer at `/mcp` endpoint
   - Graceful shutdown handling for both transports

2. **Service Layer** (`services/`)
   - `atlassian.go`: Singleton Confluence client initialization using go-atlassian library
   - `httpclient.go`: HTTP client utilities

3. **Tools Layer** (`tools/`)
   - Each tool is a separate file implementing specific Confluence operations
   - Tools register handlers with the MCP server
   - Available tools:
     - `search-page.go`: Search Confluence pages using CQL
     - `get-page.go`: Retrieve page content and metadata
     - `create-page.go`: Create new pages
     - `update-page.go`: Update existing pages
     - `get-comments.go`: Get page comments
     - `list-spaces.go`: List Confluence spaces

### MCP Integration Pattern

Each tool follows this pattern:
1. Define input/output structs with JSON tags
2. Create handler function accepting `(context, request, input) -> (*CallToolResult, error)`
3. Register tool with MCP server using `server.AddTool()`
4. Use `services.ConfluenceClient()` to get Confluence API client

### Environment Configuration

Required environment variables (set via `.env` file or directly):
- `ATLASSIAN_HOST`: Confluence instance URL (e.g., `your-domain.atlassian.net`)
- `ATLASSIAN_EMAIL`: Atlassian account email
- `ATLASSIAN_TOKEN`: Atlassian API token

## Key Dependencies

- `github.com/mark3labs/mcp-go v0.32.0`: MCP protocol implementation
- `github.com/ctreminiom/go-atlassian v1.6.1`: Confluence API client
- `github.com/joho/godotenv`: Environment variable loading

## Release Process

Uses GitHub Actions with:
- Release Please for automated changelog and version management
- GoReleaser for cross-platform binary builds
- GitHub Container Registry for Docker images