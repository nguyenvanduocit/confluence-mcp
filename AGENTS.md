# Project Summary
This repository contains a Go-based Model Context Protocol (MCP) server that integrates AI assistants with Atlassian Confluence. The server exposes several tools via MCP for searching, reading and modifying Confluence content.

## Structure
- `main.go` initializes the MCP server and registers tools.
- `services/` provides Confluence client setup and HTTP utilities.
- `tools/` defines individual tool handlers such as `search_page`, `get_page`, `create_page`, `update_page`, and `get_comments`.
- `util/` contains common helpers for error handling.

## Setup
1. Environment variables `ATLASSIAN_HOST`, `ATLASSIAN_EMAIL`, and `ATLASSIAN_TOKEN` are required.
2. Use `go run main.go --env .env --sse_port <port>` for local development or build with `go build`.
3. A `Dockerfile` and `justfile` are provided for container usage and build helpers.

## Guidelines for Codex
- Run `golangci-lint run` before committing.
- There are no tests provided.
- Pull requests should include a concise description of changes and reference relevant files.
