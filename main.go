package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/confluence-mcp/tools"
)


type CleanupFunc func()

func main() {
	envFile := flag.String("env", "", "Path to environment file (optional when environment variables are set directly)")
	streamableHttpPort := flag.String("http_port", "", "Port for streamable HTTP server. If not provided, will use stdio")
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
	if missingEnvs && *streamableHttpPort == ""{
		fmt.Println("Required environment variables missing. You must provide them via .env file or directly as environment variables.")
		fmt.Println("If using docker: docker run -e ATLASSIAN_HOST=value -e ATLASSIAN_EMAIL=value -e ATLASSIAN_TOKEN=value ...")
	}
	

	mcpServer := server.NewMCPServer(
		"Confluence Tool",
		"1.0.0",
		server.WithRecovery(),
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// Register Confluence tools
	tools.RegisterSearchPageTool(mcpServer)
	tools.RegisterGetPageTool(mcpServer)
	tools.RegisterCreatePageTool(mcpServer)
	tools.RegisterUpdatePageTool(mcpServer)
	tools.RegisterGetCommentsPageTool(mcpServer)
	tools.RegisterListSpacesTool(mcpServer)

	 // Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var cleanupFunc CleanupFunc
	
	go func() {
		if *streamableHttpPort != "" {
			log.Println("Add endpoint path http://localhost:" + *streamableHttpPort + "/mcp")
			streamableHttpServer := server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath("/mcp"),)
			cleanupFunc = func() {
				log.Println("Stopping Streamable HTTP server")
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				streamableHttpServer.Shutdown(ctx)
			}
			
			if err := streamableHttpServer.Start(fmt.Sprintf(":%s", *streamableHttpPort)); err != nil {
				log.Fatalf("Server error: %v", err)
			}
		} else {
			if err := server.ServeStdio(mcpServer); err != nil {
				cleanupFunc = func() {
					log.Println("Stopping stdio server")
				}
				log.Fatalf("Server error: %v", err)
			}
		}
	}()

	<-sigChan
	log.Println("Received signal to stop server")
	if cleanupFunc != nil {
		cleanupFunc()
	}
	log.Println("Server stopped")
	os.Exit(0)
}
