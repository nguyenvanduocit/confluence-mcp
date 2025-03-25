# Confluence Tool

A tool for interacting with Confluence API through MCP.

## Features

- Search Confluence
- Get page content
- Create new pages
- Update existing pages

## Setup

1. Clone the repository
2. Set up environment variables in `.env` file:
   ```
   ATLASSIAN_HOST=your_atlassian_host
   ATLASSIAN_EMAIL=your_email
   ATLASSIAN_TOKEN=your_token
   ```
3. Build and run the tool

## Usage

Run the tool in SSE mode:
```
just dev
```

Or build and install:
```
just build
just install
```
