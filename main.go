package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/tools"
)

func main() {
	envFile := flag.String("env", "", "Path to environment file (optional when environment variables are set directly)")
	ssePort := flag.String("sse_port", "", "Port for SSE server. If not provided, will use stdio")
	flag.Parse()

	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			fmt.Printf("Warning: Error loading env file %s: %v\n", *envFile, err)
		}
	}

	// Check required envs for Docker/production
	requiredEnvs := []string{"ATLASSIAN_HOST", "ATLASSIAN_EMAIL", "ATLASSIAN_TOKEN"}
	missingEnvs := false
	for _, env := range requiredEnvs {
		if os.Getenv(env) == "" {
			fmt.Printf("Warning: Required environment variable %s is not set\n", env)
			missingEnvs = true
		}
	}
	if missingEnvs {
		fmt.Println("Required environment variables missing. You must provide them via .env file or directly as environment variables.")
		fmt.Println("If using docker: docker run -e ATLASSIAN_HOST=value -e ATLASSIAN_EMAIL=value -e ATLASSIAN_TOKEN=value ...")
	}

	mcpServer := server.NewMCPServer(
		"Confluence Tool",
		"1.0.0",
		server.WithLogging(),
		server.WithPromptCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	// Register Confluence tools
	tools.RegisterSearchPageTool(mcpServer)
	tools.RegisterGetPageTool(mcpServer)
	tools.RegisterCreatePageTool(mcpServer)
	tools.RegisterUpdatePageTool(mcpServer)
	tools.RegisterGetCommentsPageTool(mcpServer)
	if *ssePort != "" {
		sseServer := server.NewSSEServer(mcpServer)
		if err := sseServer.Start(fmt.Sprintf(":%s", *ssePort)); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		if err := server.ServeStdio(mcpServer); err != nil {
			panic(fmt.Sprintf("Server error: %v", err))
		}
	}
}
