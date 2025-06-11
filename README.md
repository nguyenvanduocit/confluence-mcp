# Confluence MCP

A Go-based Model Context Protocol (MCP) server for integrating AI assistants with Atlassian Confluence. This tool provides a seamless interface for interacting with the Confluence API through the standardized MCP, enabling AI models to search, retrieve, create, and update Confluence content.

![](/assets//thumbnail.webp)

## Features

- Search Confluence pages and spaces
- Get page details and content
- Create new pages and spaces
- Update existing pages
- Manage page permissions and metadata
- Get page comments
- List spaces

## Installation

There are several ways to install the Confluence MCP:

### Option 1: Download from GitHub Releases

1. Visit the [GitHub Releases](https://github.com/nguyenvanduocit/confluence-mcp/releases) page
2. Download the binary for your platform:
   - `confluence-mcp_linux_amd64` for Linux
   - `confluence-mcp_darwin_amd64` for macOS
   - `confluence-mcp_windows_amd64.exe` for Windows
3. Make the binary executable (Linux/macOS):
   ```bash
   chmod +x confluence-mcp_*
   ```
4. Move it to your PATH (Linux/macOS):
   ```bash
   sudo mv confluence-mcp_* /usr/local/bin/confluence-mcp
   ```

### Option 2: Go install
```bash
go install github.com/nguyenvanduocit/confluence-mcp@latest
```

### Option 3: Docker

#### Using Docker directly
1. Pull the pre-built image from GitHub Container Registry:
   ```bash
   docker pull ghcr.io/nguyenvanduocit/confluence-mcp:latest
   ```

2. Or build the Docker image locally:
   ```bash
   docker build -t confluence-mcp .
   ```

## Configuration

### Environment Variables
The following environment variables are required for authentication:
```
ATLASSIAN_HOST=your_confluence_host
ATLASSIAN_EMAIL=your_email
ATLASSIAN_TOKEN=your_token
```

You can set these directly in environment variables or through a `.env` file for local development.

### Transport Methods

The Confluence MCP supports two transport methods:

#### 1. Standard I/O (stdio) - Default
This is the default transport method used by most MCP clients like Claude Desktop and Cursor.

#### 2. Streamable HTTP Server
For HTTP-based integrations, you can run the server with HTTP transport using the `--http_port` flag.

## Usage Examples

### With Claude Desktop / Cursor (stdio transport)

Add to your MCP configuration file:

```json
{
  "mcpServers": {
    "confluence": {
      "command": "/path/to/confluence-mcp",
      "args": ["-env", "/path/to/.env"]
    }
  }
}
```

### With Docker (stdio transport)

```json
{
  "mcpServers": {
    "confluence": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-e", "ATLASSIAN_HOST=your_confluence_instance.atlassian.net",
        "-e", "ATLASSIAN_EMAIL=your_email@example.com",
        "-e", "ATLASSIAN_TOKEN=your_atlassian_api_token",
        "ghcr.io/nguyenvanduocit/confluence-mcp:latest"
      ]
    }
  }
}
```

### With HTTP Transport

For HTTP-based integrations, run the server with:

```bash
confluence-mcp --http_port 8080 --env .env
```

This will start the server at `http://localhost:8080/mcp`

Or with Docker:

```bash
docker run -p 8080:8080 \
  -e ATLASSIAN_HOST=your_confluence_instance.atlassian.net \
  -e ATLASSIAN_EMAIL=your_email@example.com \
  -e ATLASSIAN_TOKEN=your_atlassian_api_token \
  ghcr.io/nguyenvanduocit/confluence-mcp:latest \
  --http_port 8080
```

## Available Tools

- `search_page` - Search pages in Confluence using CQL
- `get_page` - Get Confluence page content and metadata
- `create_page` - Create new Confluence pages
- `update_page` - Update existing Confluence pages
- `get_comments` - Get comments from a Confluence page
- `list_spaces` - List Confluence spaces

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
