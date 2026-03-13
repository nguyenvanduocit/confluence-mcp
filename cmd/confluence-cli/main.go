package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/joho/godotenv"
	"github.com/nguyenvanduocit/confluence-mcp/services"
	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "search-page":
		runSearchPage(os.Args[2:])
	case "get-page":
		runGetPage(os.Args[2:])
	case "create-page":
		runCreatePage(os.Args[2:])
	case "update-page":
		runUpdatePage(os.Args[2:])
	case "get-comments":
		runGetComments(os.Args[2:])
	case "list-spaces":
		runListSpaces(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `confluence-cli - Confluence CLI tool

Usage:
  confluence-cli <command> [flags]

Commands:
  search-page    Search Confluence pages using CQL
  get-page       Get a Confluence page by ID
  create-page    Create a new Confluence page
  update-page    Update an existing Confluence page
  get-comments   Get comments for a Confluence page
  list-spaces    List Confluence spaces

Global Flags:
  --env string     Path to .env file
  --output string  Output format: text|json (default "text")

Run 'confluence-cli <command> --help' for command-specific flags.
`)
}

func loadEnv(envFile string) {
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not load env file %s: %v\n", envFile, err)
		}
	}
}

func outputResult(v interface{}, format string) {
	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(v); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode JSON: %v\n", err)
			os.Exit(1)
		}
	default:
		data, err := yaml.Marshal(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode output: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(data))
	}
}

func runSearchPage(args []string) {
	fs := flag.NewFlagSet("search-page", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	query := fs.String("query", "", "CQL query (required)")
	output := fs.String("output", "text", "Output format: text|json")
	fs.Parse(args)

	loadEnv(*env)

	if *query == "" {
		fmt.Fprintln(os.Stderr, "Error: --query is required")
		fs.Usage()
		os.Exit(1)
	}

	client, err := services.ConfluenceClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	options := &models.SearchContentOptions{
		Limit: 5,
	}

	contents, response, err := client.Search.Content(context.Background(), *query, options)
	if err != nil {
		if response != nil {
			fmt.Fprintf(os.Stderr, "search failed: %s (endpoint: %s)\n", response.Bytes.String(), response.Endpoint)
		} else {
			fmt.Fprintf(os.Stderr, "search failed: %v\n", err)
		}
		os.Exit(1)
	}

	type SearchResult struct {
		Title        string `json:"title" yaml:"title"`
		ID           string `json:"id" yaml:"id"`
		Type         string `json:"type" yaml:"type"`
		Link         string `json:"link" yaml:"link"`
		LastModified string `json:"last_modified" yaml:"last_modified"`
		Excerpt      string `json:"excerpt" yaml:"excerpt"`
	}
	type SearchPageOutput struct {
		Query       string         `json:"query" yaml:"query"`
		Results     []SearchResult `json:"results" yaml:"results"`
		ResultCount int            `json:"result_count" yaml:"result_count"`
		Message     string         `json:"message" yaml:"message"`
	}

	out := SearchPageOutput{
		Query:       *query,
		Results:     make([]SearchResult, 0, len(contents.Results)),
		ResultCount: len(contents.Results),
	}

	if len(contents.Results) == 0 {
		out.Message = "No results found for the search query"
	} else {
		for _, content := range contents.Results {
			out.Results = append(out.Results, SearchResult{
				Title:        content.Content.Title,
				ID:           content.Content.ID,
				Type:         content.Content.Type,
				Link:         content.Content.Links.Self,
				LastModified: content.LastModified,
				Excerpt:      content.Excerpt,
			})
		}
		out.Message = fmt.Sprintf("Found %d results for query: %s", len(contents.Results), *query)
	}

	outputResult(out, *output)
}

func runGetPage(args []string) {
	fs := flag.NewFlagSet("get-page", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	id := fs.String("id", "", "Confluence page ID (required)")
	output := fs.String("output", "text", "Output format: text|json")
	fs.Parse(args)

	loadEnv(*env)

	if *id == "" {
		fmt.Fprintln(os.Stderr, "Error: --id is required")
		fs.Usage()
		os.Exit(1)
	}

	client, err := services.ConfluenceClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.Background()
	expandParams := []string{"body.storage", "body.view", "body"}
	content, response, err := client.Content.Get(ctx, *id, expandParams, 0)
	if err != nil {
		if response != nil {
			fmt.Fprintf(os.Stderr, "failed to get page: %s (endpoint: %s)\n", response.Bytes.String(), response.Endpoint)
		} else {
			fmt.Fprintf(os.Stderr, "failed to get page: %v\n", err)
		}
		os.Exit(1)
	}

	var htmlContent string
	if content.Body == nil {
		htmlContent = "Page body is nil - no content available"
	} else if content.Body.Storage != nil && content.Body.Storage.Value != "" {
		htmlContent = content.Body.Storage.Value
	} else if content.Body.View != nil && content.Body.View.Value != "" {
		htmlContent = content.Body.View.Value
	} else {
		htmlContent = "No content available in either storage or view format"
	}

	var versionNumber int
	if content.Version != nil {
		versionNumber = content.Version.Number
	}

	type PageInfo struct {
		Title   string `json:"title" yaml:"title"`
		ID      string `json:"id" yaml:"id"`
		Version int    `json:"version" yaml:"version"`
	}
	type GetPageOutput struct {
		Title          string     `json:"title" yaml:"title"`
		ID             string     `json:"id" yaml:"id"`
		Version        int        `json:"version" yaml:"version"`
		Type           string     `json:"type" yaml:"type"`
		Content        string     `json:"content" yaml:"content"`
		DirectChildren []PageInfo `json:"direct_children,omitempty" yaml:"direct_children,omitempty"`
		AllDescendants []PageInfo `json:"all_descendants,omitempty" yaml:"all_descendants,omitempty"`
		Message        string     `json:"message" yaml:"message"`
	}

	out := GetPageOutput{
		Title:   content.Title,
		ID:      content.ID,
		Version: versionNumber,
		Type:    content.Type,
		Content: htmlContent,
	}

	childPages, childResponse, err := client.Content.ChildrenDescendant.ChildrenByType(
		ctx, *id, "page", 0, []string{"title", "id", "version"}, 0, 100,
	)
	if err == nil && childResponse != nil && childPages != nil && len(childPages.Results) > 0 {
		out.DirectChildren = make([]PageInfo, 0, len(childPages.Results))
		for _, childPage := range childPages.Results {
			var childVersion int
			if childPage.Version != nil {
				childVersion = childPage.Version.Number
			}
			out.DirectChildren = append(out.DirectChildren, PageInfo{
				Title:   childPage.Title,
				ID:      childPage.ID,
				Version: childVersion,
			})
		}
	}

	descendants, descendantsResponse, err := client.Content.ChildrenDescendant.DescendantsByType(
		ctx, *id, "page", "all", []string{"title", "id", "version"}, 0, 100,
	)
	if err == nil && descendantsResponse != nil && descendants != nil && len(descendants.Results) > 0 {
		directChildren := make(map[string]bool)
		if childPages != nil {
			for _, child := range childPages.Results {
				directChildren[child.ID] = true
			}
		}
		var nonDirect []*models.ContentScheme
		for _, d := range descendants.Results {
			if !directChildren[d.ID] {
				nonDirect = append(nonDirect, d)
			}
		}
		if len(nonDirect) > 0 {
			out.AllDescendants = make([]PageInfo, 0, len(nonDirect))
			for _, d := range nonDirect {
				var dv int
				if d.Version != nil {
					dv = d.Version.Number
				}
				out.AllDescendants = append(out.AllDescendants, PageInfo{
					Title:   d.Title,
					ID:      d.ID,
					Version: dv,
				})
			}
		}
	}

	out.Message = fmt.Sprintf("Page retrieved successfully with %d direct children and %d other descendants",
		len(out.DirectChildren), len(out.AllDescendants))

	outputResult(out, *output)
}

func runCreatePage(args []string) {
	fs := flag.NewFlagSet("create-page", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	space := fs.String("space", "", "Space key (required)")
	title := fs.String("title", "", "Page title (required)")
	content := fs.String("content", "", "Page content in storage format XHTML (required)")
	parentID := fs.String("parent-id", "", "Parent page ID (optional)")
	output := fs.String("output", "text", "Output format: text|json")
	fs.Parse(args)

	loadEnv(*env)

	if *space == "" || *title == "" || *content == "" {
		fmt.Fprintln(os.Stderr, "Error: --space, --title, and --content are required")
		fs.Usage()
		os.Exit(1)
	}

	client, err := services.ConfluenceClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	payload := &models.ContentScheme{
		Type:  "page",
		Title: *title,
		Space: &models.SpaceScheme{
			Key: *space,
		},
		Body: &models.BodyScheme{
			Storage: &models.BodyNodeScheme{
				Value:          *content,
				Representation: "storage",
			},
		},
	}

	if *parentID != "" {
		payload.Ancestors = []*models.ContentScheme{{ID: *parentID}}
	}

	newPage, response, err := client.Content.Create(context.Background(), payload)
	if err != nil {
		if response != nil {
			fmt.Fprintf(os.Stderr, "failed to create page: %s (endpoint: %s)\n", response.Bytes.String(), response.Endpoint)
		} else {
			fmt.Fprintf(os.Stderr, "failed to create page: %v\n", err)
		}
		os.Exit(1)
	}

	var versionNumber int
	if newPage.Version != nil {
		versionNumber = newPage.Version.Number
	}
	var selfLink string
	if newPage.Links != nil {
		selfLink = newPage.Links.Self
	}

	type CreatePageOutput struct {
		Success bool   `json:"success" yaml:"success"`
		Title   string `json:"title" yaml:"title"`
		ID      string `json:"id" yaml:"id"`
		Version int    `json:"version" yaml:"version"`
		Link    string `json:"link" yaml:"link"`
		Message string `json:"message" yaml:"message"`
	}

	out := CreatePageOutput{
		Success: true,
		Title:   newPage.Title,
		ID:      newPage.ID,
		Version: versionNumber,
		Link:    selfLink,
		Message: fmt.Sprintf("Page created successfully!\nTitle: %s\nID: %s\nVersion: %d\nLink: %s",
			newPage.Title, newPage.ID, versionNumber, selfLink),
	}

	outputResult(out, *output)
}

func runUpdatePage(args []string) {
	fs := flag.NewFlagSet("update-page", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	id := fs.String("id", "", "Page ID (required)")
	title := fs.String("title", "", "New page title (required)")
	content := fs.String("content", "", "New page content in storage format XHTML (required)")
	version := fs.String("version", "", "Version number override (optional)")
	output := fs.String("output", "text", "Output format: text|json")
	fs.Parse(args)

	loadEnv(*env)

	if *id == "" || *title == "" || *content == "" {
		fmt.Fprintln(os.Stderr, "Error: --id, --title, and --content are required")
		fs.Usage()
		os.Exit(1)
	}

	client, err := services.ConfluenceClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.Background()

	currentPage, response, err := client.Content.Get(ctx, *id, []string{"version"}, 0)
	if err != nil {
		if response != nil {
			fmt.Fprintf(os.Stderr, "failed to get current page: %s (endpoint: %s)\n", response.Bytes.String(), response.Endpoint)
		} else {
			fmt.Fprintf(os.Stderr, "failed to get current page: %v\n", err)
		}
		os.Exit(1)
	}

	payload := &models.ContentScheme{
		ID:    *id,
		Type:  "page",
		Title: currentPage.Title,
	}

	if currentPage.Version != nil {
		payload.Version = &models.ContentVersionScheme{
			Number: currentPage.Version.Number + 1,
		}
	} else {
		payload.Version = &models.ContentVersionScheme{Number: 1}
	}

	if *title != "" {
		payload.Title = *title
	}

	if *content != "" {
		payload.Body = &models.BodyScheme{
			Storage: &models.BodyNodeScheme{
				Value:          *content,
				Representation: "storage",
			},
		}
	}

	if *version != "" {
		v, err := strconv.Atoi(*version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid --version: %v\n", err)
			os.Exit(1)
		}
		payload.Version.Number = v
	}

	updatedPage, response, err := client.Content.Update(ctx, *id, payload)
	if err != nil {
		if response != nil {
			fmt.Fprintf(os.Stderr, "failed to update page: %s (endpoint: %s)\n", response.Bytes.String(), response.Endpoint)
		} else {
			fmt.Fprintf(os.Stderr, "failed to update page: %v\n", err)
		}
		os.Exit(1)
	}

	var versionNumber int
	if updatedPage.Version != nil {
		versionNumber = updatedPage.Version.Number
	}
	var selfLink string
	if updatedPage.Links != nil {
		selfLink = updatedPage.Links.Self
	}

	type UpdatePageOutput struct {
		Success bool   `json:"success" yaml:"success"`
		Title   string `json:"title" yaml:"title"`
		ID      string `json:"id" yaml:"id"`
		Version int    `json:"version" yaml:"version"`
		Link    string `json:"link" yaml:"link"`
		Message string `json:"message" yaml:"message"`
	}

	out := UpdatePageOutput{
		Success: true,
		Title:   updatedPage.Title,
		ID:      updatedPage.ID,
		Version: versionNumber,
		Link:    selfLink,
		Message: fmt.Sprintf("Page updated successfully!\nTitle: %s\nID: %s\nVersion: %d\nLink: %s",
			updatedPage.Title, updatedPage.ID, versionNumber, selfLink),
	}

	outputResult(out, *output)
}

func runGetComments(args []string) {
	fs := flag.NewFlagSet("get-comments", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	id := fs.String("id", "", "Confluence page ID (required)")
	output := fs.String("output", "text", "Output format: text|json")
	fs.Parse(args)

	loadEnv(*env)

	if *id == "" {
		fmt.Fprintln(os.Stderr, "Error: --id is required")
		fs.Usage()
		os.Exit(1)
	}

	client, err := services.ConfluenceClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	maxResults := 50
	comments, response, err := client.Content.Comment.Gets(
		context.Background(), *id, []string{}, []string{}, 0, maxResults,
	)
	if err != nil {
		if response != nil {
			fmt.Fprintf(os.Stderr, "failed to get comments: %s (endpoint: %s)\n", response.Bytes.String(), response.Endpoint)
		} else {
			fmt.Fprintf(os.Stderr, "failed to get comments: %v\n", err)
		}
		os.Exit(1)
	}

	type CommentInfo struct {
		ID      string `json:"id" yaml:"id"`
		Title   string `json:"title" yaml:"title"`
		Status  string `json:"status" yaml:"status"`
		Author  string `json:"author,omitempty" yaml:"author,omitempty"`
		Created string `json:"created,omitempty" yaml:"created,omitempty"`
		Content string `json:"content,omitempty" yaml:"content,omitempty"`
	}
	type GetCommentsOutput struct {
		PageID      string        `json:"page_id" yaml:"page_id"`
		Comments    []CommentInfo `json:"comments" yaml:"comments"`
		TotalCount  int           `json:"total_count" yaml:"total_count"`
		CurrentPage int           `json:"current_page" yaml:"current_page"`
		Message     string        `json:"message" yaml:"message"`
	}

	out := GetCommentsOutput{
		PageID:      *id,
		Comments:    make([]CommentInfo, 0, len(comments.Results)),
		TotalCount:  comments.Size,
		CurrentPage: 1,
	}

	if len(comments.Results) == 0 {
		out.Message = "No comments found."
	} else {
		for _, comment := range comments.Results {
			ci := CommentInfo{
				ID:     comment.ID,
				Title:  comment.Title,
				Status: comment.Status,
			}
			if comment.Version != nil && comment.Version.By != nil {
				ci.Author = comment.Version.By.DisplayName
				ci.Created = comment.Version.When
			}
			if comment.Body != nil && comment.Body.View != nil {
				ci.Content = comment.Body.View.Value
			}
			out.Comments = append(out.Comments, ci)
		}
		out.Message = fmt.Sprintf("Found %d comments (page 1)", len(comments.Results))
	}

	outputResult(out, *output)
}

func runListSpaces(args []string) {
	fs := flag.NewFlagSet("list-spaces", flag.ExitOnError)
	env := fs.String("env", "", "Path to .env file")
	output := fs.String("output", "text", "Output format: text|json")
	fs.Parse(args)

	loadEnv(*env)

	client, err := services.ConfluenceClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	spaces, response, err := client.Space.Gets(context.Background(), nil, 0, 100)
	if err != nil {
		if response != nil {
			fmt.Fprintf(os.Stderr, "failed to list spaces: %s (endpoint: %s)\n", response.Bytes.String(), response.Endpoint)
		} else {
			fmt.Fprintf(os.Stderr, "failed to list spaces: %v\n", err)
		}
		os.Exit(1)
	}

	type SpaceInfo struct {
		Key    string `json:"key" yaml:"key"`
		ID     int    `json:"id" yaml:"id"`
		Name   string `json:"name" yaml:"name"`
		Type   string `json:"type" yaml:"type"`
		Status string `json:"status" yaml:"status"`
		Link   string `json:"link" yaml:"link"`
	}
	type ListSpacesOutput struct {
		Spaces     []SpaceInfo `json:"spaces" yaml:"spaces"`
		SpaceCount int         `json:"space_count" yaml:"space_count"`
		Message    string      `json:"message" yaml:"message"`
	}

	out := ListSpacesOutput{
		Spaces:     make([]SpaceInfo, 0, len(spaces.Results)),
		SpaceCount: len(spaces.Results),
	}

	if len(spaces.Results) == 0 {
		out.Message = "No spaces found"
	} else {
		for _, space := range spaces.Results {
			var link string
			if space.Links != nil {
				link = space.Links.Self
			}
			out.Spaces = append(out.Spaces, SpaceInfo{
				Key:    space.Key,
				ID:     space.ID,
				Name:   space.Name,
				Type:   space.Type,
				Status: space.Status,
				Link:   link,
			})
		}
		out.Message = fmt.Sprintf("Found %d spaces", len(spaces.Results))
	}

	outputResult(out, *output)
}
