package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sgaunet/gitlab-mcp/internal/app"
	"github.com/sgaunet/gitlab-mcp/internal/logger"
)

// Version information injected at build time.
var version = "dev"

const (
	defaultLimit = 100
)

// Error variables for static errors.
var (
	ErrInvalidStateValue = errors.New("state must be 'opened' or 'closed'")
)

// setupListIssuesTool creates and registers the list_issues tool.
func setupListIssuesTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	listIssuesTool := mcp.NewTool("list_issues",
		mcp.WithDescription("List issues for a GitLab project by project path"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("GitLab project path (e.g., 'namespace/project-name')"),
		),
		mcp.WithString("state",
			mcp.Description("Filter by issue state: opened, closed, or all (default: opened)"),
		),
		mcp.WithString("labels",
			mcp.Description("Comma-separated list of labels to filter by"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of issues to return (default: 100, max: 100)"),
		),
	)

	s.AddTool(listIssuesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received list_issues tool request", "args", args)

		// Extract project_path
		projectPath, ok := args["project_path"].(string)
		if !ok || projectPath == "" {
			debugLogger.Error("project_path is not a valid string", "value", args["project_path"])
			return mcp.NewToolResultError("project_path must be a non-empty string"), nil
		}

		// Extract optional parameters
		opts := &app.ListIssuesOptions{
			State: "opened", // default
			Limit: defaultLimit, // default
		}

		if state, ok := args["state"].(string); ok && state != "" {
			opts.State = state
		}

		if labels, ok := args["labels"].(string); ok && labels != "" {
			opts.Labels = labels
		}

		if limitFloat, ok := args["limit"].(float64); ok {
			opts.Limit = int(limitFloat)
		}

		debugLogger.Debug("Processing list_issues request", "project_path", projectPath, "opts", opts)

		// Call the app method
		issues, err := appInstance.ListProjectIssues(projectPath, opts)
		if err != nil {
			debugLogger.Error("Failed to list project issues", "error", err, "project_path", projectPath)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list project issues: %v", err)), nil
		}

		// Convert issues to JSON
		jsonData, err := json.Marshal(issues)
		if err != nil {
			debugLogger.Error("Failed to marshal issues to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format issues response"), nil
		}

		debugLogger.Info("Successfully retrieved project issues", "count", len(issues), "project_path", projectPath)
		return mcp.NewToolResultText(string(jsonData)), nil
	})
}

// setupCreateIssueTool creates and registers the create_issues tool.
func setupCreateIssueTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	createIssueTool := mcp.NewTool("create_issues",
		mcp.WithDescription("Create a new issue for a GitLab project by project path"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("GitLab project path (e.g., 'namespace/project-name')"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Issue title"),
		),
		mcp.WithString("description",
			mcp.Description("Issue description"),
		),
		mcp.WithArray("labels",
			mcp.Description("Array of labels to assign to the issue"),
		),
		mcp.WithArray("assignees",
			mcp.Description("Array of user IDs to assign to the issue"),
		),
	)

	s.AddTool(createIssueTool, handleCreateIssueRequest(appInstance, debugLogger))
}

// handleCreateIssueRequest handles the create_issues tool request.
func handleCreateIssueRequest(
	appInstance *app.App, 
	debugLogger *slog.Logger,
) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received create_issues tool request", "args", args)

		// Extract project_path
		projectPath, ok := args["project_path"].(string)
		if !ok || projectPath == "" {
			debugLogger.Error("project_path is not a valid string", "value", args["project_path"])
			return mcp.NewToolResultError("project_path must be a non-empty string"), nil
		}

		// Extract title (required)
		title, ok := args["title"].(string)
		if !ok || title == "" {
			debugLogger.Error("title is missing or not a string", "value", args["title"])
			return mcp.NewToolResultError("title must be a non-empty string"), nil
		}

		// Extract options
		opts := extractCreateIssueOptions(args, title)

		debugLogger.Debug("Processing create_issues request", "project_path", projectPath, "title", title)

		// Call the app method
		issue, err := appInstance.CreateProjectIssue(projectPath, opts)
		if err != nil {
			debugLogger.Error("Failed to create issue", "error", err, "project_path", projectPath, "title", title)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create issue: %v", err)), nil
		}

		// Convert issue to JSON
		jsonData, err := json.Marshal(issue)
		if err != nil {
			debugLogger.Error("Failed to marshal issue to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format issue response"), nil
		}

		debugLogger.Info("Successfully created issue", 
			"id", issue.ID, 
			"iid", issue.IID, 
			"project_path", projectPath, 
			"title", issue.Title)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// extractCreateIssueOptions extracts create issue options from arguments.
func extractCreateIssueOptions(args map[string]interface{}, title string) *app.CreateIssueOptions {
	opts := &app.CreateIssueOptions{
		Title: title,
	}

	// Extract optional description
	if description, ok := args["description"].(string); ok {
		opts.Description = description
	}

	// Extract optional labels
	if labelsInterface, ok := args["labels"].([]interface{}); ok {
		labels := make([]string, 0, len(labelsInterface))
		for _, label := range labelsInterface {
			if labelStr, ok := label.(string); ok {
				labels = append(labels, labelStr)
			}
		}
		opts.Labels = labels
	}

	// Extract optional assignees
	if assigneesInterface, ok := args["assignees"].([]interface{}); ok {
		assignees := make([]int, 0, len(assigneesInterface))
		for _, assignee := range assigneesInterface {
			if assigneeFloat, ok := assignee.(float64); ok {
				assignees = append(assignees, int(assigneeFloat))
			}
		}
		opts.Assignees = assignees
	}

	return opts
}

// setupUpdateIssueTool creates and registers the update_issues tool.
func setupUpdateIssueTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	updateIssueTool := mcp.NewTool("update_issues",
		mcp.WithDescription("Update an existing issue for a GitLab project by project path"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("GitLab project path (e.g., 'namespace/project-name')"),
		),
		mcp.WithNumber("issue_iid",
			mcp.Required(),
			mcp.Description("Issue internal ID (IID) to update"),
		),
		mcp.WithString("title",
			mcp.Description("Updated issue title"),
		),
		mcp.WithString("description",
			mcp.Description("Updated issue description"),
		),
		mcp.WithString("state",
			mcp.Description("Issue state: 'opened' or 'closed'"),
		),
		mcp.WithArray("labels",
			mcp.Description("Array of labels to assign to the issue"),
		),
		mcp.WithArray("assignees",
			mcp.Description("Array of user IDs to assign to the issue"),
		),
	)

	s.AddTool(updateIssueTool, handleUpdateIssueRequest(appInstance, debugLogger))
}

// handleUpdateIssueRequest handles the update_issues tool request.
func handleUpdateIssueRequest(
	appInstance *app.App, 
	debugLogger *slog.Logger,
) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received update_issues tool request", "args", args)

		// Extract project_path
		projectPath, ok := args["project_path"].(string)
		if !ok || projectPath == "" {
			debugLogger.Error("project_path is not a valid string", "value", args["project_path"])
			return mcp.NewToolResultError("project_path must be a non-empty string"), nil
		}

		// Extract issue_iid (required)
		issueIIDFloat, ok := args["issue_iid"].(float64)
		if !ok {
			debugLogger.Error("issue_iid is missing or not a number", "value", args["issue_iid"])
			return mcp.NewToolResultError("issue_iid must be a number"), nil
		}
		issueIID := int(issueIIDFloat)

		// Extract options
		opts, err := extractUpdateIssueOptions(args, debugLogger)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		debugLogger.Debug("Processing update_issues request", "project_path", projectPath, "issue_iid", issueIID)

		// Call the app method
		issue, err := appInstance.UpdateProjectIssue(projectPath, issueIID, opts)
		if err != nil {
			debugLogger.Error("Failed to update issue", "error", err, "project_path", projectPath, "issue_iid", issueIID)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update issue: %v", err)), nil
		}

		// Convert issue to JSON
		jsonData, err := json.Marshal(issue)
		if err != nil {
			debugLogger.Error("Failed to marshal issue to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format issue response"), nil
		}

		debugLogger.Info("Successfully updated issue", "id", issue.ID, "iid", issue.IID, "project_path", projectPath)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// extractUpdateIssueOptions extracts update issue options from arguments.
func extractUpdateIssueOptions(args map[string]interface{}, debugLogger *slog.Logger) (*app.UpdateIssueOptions, error) {
	opts := &app.UpdateIssueOptions{}

	// Extract basic string fields
	extractUpdateStringFields(args, opts)

	// Extract and validate state
	if err := extractUpdateState(args, opts, debugLogger); err != nil {
		return nil, err
	}

	// Extract array fields
	extractUpdateArrayFields(args, opts)

	return opts, nil
}

// extractUpdateStringFields extracts string fields for update options.
func extractUpdateStringFields(args map[string]interface{}, opts *app.UpdateIssueOptions) {
	if title, ok := args["title"].(string); ok && title != "" {
		opts.Title = title
	}

	if description, ok := args["description"].(string); ok {
		opts.Description = description
	}
}

// extractUpdateState extracts and validates the state field.
func extractUpdateState(args map[string]interface{}, opts *app.UpdateIssueOptions, debugLogger *slog.Logger) error {
	if state, ok := args["state"].(string); ok && state != "" {
		if state != "opened" && state != "closed" {
			debugLogger.Error("invalid state value", "state", state)
			return ErrInvalidStateValue
		}
		opts.State = state
	}
	return nil
}

// extractUpdateArrayFields extracts array fields for update options.
func extractUpdateArrayFields(args map[string]interface{}, opts *app.UpdateIssueOptions) {
	// Extract optional labels
	if labelsInterface, ok := args["labels"].([]interface{}); ok {
		labels := make([]string, 0, len(labelsInterface))
		for _, label := range labelsInterface {
			if labelStr, ok := label.(string); ok {
				labels = append(labels, labelStr)
			}
		}
		opts.Labels = labels
	}

	// Extract optional assignees
	if assigneesInterface, ok := args["assignees"].([]interface{}); ok {
		assignees := make([]int, 0, len(assigneesInterface))
		for _, assignee := range assigneesInterface {
			if assigneeFloat, ok := assignee.(float64); ok {
				assignees = append(assignees, int(assigneeFloat))
			}
		}
		opts.Assignees = assignees
	}
}

// setupListLabelsTool creates and registers the list_labels tool.
func setupListLabelsTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	listLabelsTool := mcp.NewTool("list_labels",
		mcp.WithDescription("List labels for a GitLab project by project path"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("GitLab project path (e.g., 'namespace/project-name')"),
		),
		mcp.WithBoolean("with_counts",
			mcp.Description("Include issue and merge request counts (default: false)"),
		),
		mcp.WithBoolean("include_ancestor_groups",
			mcp.Description("Include labels from ancestor groups (default: false)"),
		),
		mcp.WithString("search",
			mcp.Description("Filter labels by search keyword"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of labels to return (default: 100, max: 100)"),
		),
	)

	s.AddTool(listLabelsTool, handleListLabelsRequest(appInstance, debugLogger))
}

// handleListLabelsRequest handles the list_labels tool request.
func handleListLabelsRequest(
	appInstance *app.App, 
	debugLogger *slog.Logger,
) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received list_labels tool request", "args", args)

		// Extract project_path
		projectPath, ok := args["project_path"].(string)
		if !ok || projectPath == "" {
			debugLogger.Error("project_path is not a valid string", "value", args["project_path"])
			return mcp.NewToolResultError("project_path must be a non-empty string"), nil
		}

		// Extract optional parameters
		opts := extractListLabelsOptions(args)

		debugLogger.Debug("Processing list_labels request", "project_path", projectPath, "opts", opts)

		// Call the app method
		labels, err := appInstance.ListProjectLabels(projectPath, opts)
		if err != nil {
			debugLogger.Error("Failed to list project labels", "error", err, "project_path", projectPath)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list project labels: %v", err)), nil
		}

		// Convert labels to JSON
		jsonData, err := json.Marshal(labels)
		if err != nil {
			debugLogger.Error("Failed to marshal labels to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format labels response"), nil
		}

		debugLogger.Info("Successfully retrieved project labels", "count", len(labels), "project_path", projectPath)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// extractListLabelsOptions extracts list labels options from arguments.
func extractListLabelsOptions(args map[string]interface{}) *app.ListLabelsOptions {
	opts := &app.ListLabelsOptions{
		WithCounts:            false, // default
		IncludeAncestorGroups: false, // default
		Limit:                 defaultLimit, // default
	}

	if withCounts, ok := args["with_counts"].(bool); ok {
		opts.WithCounts = withCounts
	}

	if includeAncestorGroups, ok := args["include_ancestor_groups"].(bool); ok {
		opts.IncludeAncestorGroups = includeAncestorGroups
	}

	if search, ok := args["search"].(string); ok && search != "" {
		opts.Search = search
	}

	if limitFloat, ok := args["limit"].(float64); ok {
		opts.Limit = int(limitFloat)
	}

	return opts
}

func printHelp() {
	fmt.Printf(`GitLab MCP Server %s

A Model Context Protocol (MCP) server that provides GitLab integration tools for Claude Code.

USAGE:
    gitlab-mcp [OPTIONS]

OPTIONS:
    -h, --help     Show this help message
    -v, --version  Show version information

ENVIRONMENT VARIABLES:
    GITLAB_TOKEN   GitLab API personal access token (required)
    GITLAB_URI     GitLab instance URI (default: https://gitlab.com/)

DESCRIPTION:
    This MCP server provides the following tools for GitLab integration:
    
    • list_issues     - List issues for a GitLab project
    • create_issues   - Create new issues with metadata
    • update_issues   - Update existing issues
    • list_labels     - List project labels with filtering
    
    The server communicates via JSON-RPC 2.0 over stdin/stdout and is designed
    to be used with Claude Code's MCP architecture.

EXAMPLES:
    # Start the MCP server (typically called by Claude Code)
    gitlab-mcp
    
    # Show help
    gitlab-mcp -h
    
    # Show version
    gitlab-mcp -v

For more information, visit: https://github.com/sgaunet/gitlab-mcp
`, version)
}

func main() {
	// Parse command line flags
	var (
		showHelp    = flag.Bool("h", false, "Show help message")
		showHelpLong = flag.Bool("help", false, "Show help message")
		showVersion = flag.Bool("v", false, "Show version information") 
		showVersionLong = flag.Bool("version", false, "Show version information")
	)
	
	flag.Parse()
	
	// Handle help flags
	if *showHelp || *showHelpLong {
		printHelp()
		os.Exit(0)
	}
	
	// Handle version flags
	if *showVersion || *showVersionLong {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}

	// Initialize the app
	appInstance, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// Set debug logger
	debugLogger := logger.NewLogger("debug")
	appInstance.SetLogger(debugLogger)
	
	debugLogger.Info("Starting GitLab MCP Server", "version", version)

	// Validate connection
	if err := appInstance.ValidateConnection(); err != nil {
		log.Fatalf("Failed to validate GitLab connection: %v", err)
	}

	// Create MCP server
	s := server.NewMCPServer(
		"GitLab MCP Server",
		version,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	// Create and register list_issues tool
	setupListIssuesTool(s, appInstance, debugLogger)

	// Create and register create_issues tool
	setupCreateIssueTool(s, appInstance, debugLogger)

	// Create and register update_issues tool
	setupUpdateIssueTool(s, appInstance, debugLogger)

	// Create and register list_labels tool
	setupListLabelsTool(s, appInstance, debugLogger)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
