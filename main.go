package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sgaunet/gitlab-mcp/internal/app"
	"github.com/sgaunet/gitlab-mcp/internal/logger"
)

func main() {
	// Initialize the app
	appInstance, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// Set debug logger
	debugLogger := logger.NewLogger("debug")
	appInstance.SetLogger(debugLogger)
	
	debugLogger.Info("Starting GitLab MCP Server", "version", "1.0.0")

	// Validate connection
	if err := appInstance.ValidateConnection(); err != nil {
		log.Fatalf("Failed to validate GitLab connection: %v", err)
	}

	// Create MCP server
	s := server.NewMCPServer(
		"GitLab MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	// Create get_project_id tool
	getProjectIDTool := mcp.NewTool("get_project_id",
		mcp.WithDescription("Get GitLab project ID from git remote repository URL"),
		mcp.WithString("remote_url",
			mcp.Required(),
			mcp.Description("Git remote repository URL (SSH or HTTPS)"),
		),
	)

	// Add tool handler
	s.AddTool(getProjectIDTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received MCP tool request", "args", args)

		remoteURL, ok := args["remote_url"].(string)
		if !ok {
			debugLogger.Error("remote_url is not a string", "value", args["remote_url"])
			return mcp.NewToolResultError("remote_url must be a string"), nil
		}

		debugLogger.Debug("Processing MCP tool request", "remote_url", remoteURL)
		projectID, err := appInstance.GetProjectID(remoteURL)
		if err != nil {
			debugLogger.Error("Failed to get project ID", "error", err, "remote_url", remoteURL)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get project ID: %v", err)), nil
		}

		debugLogger.Info("Successfully retrieved project ID", "id", projectID, "remote_url", remoteURL)
		return mcp.NewToolResultText(fmt.Sprintf("%d", projectID)), nil
	})

	// Create list_issues tool
	listIssuesTool := mcp.NewTool("list_issues",
		mcp.WithDescription("List issues for a GitLab project by project ID"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("GitLab project ID"),
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

	// Add list_issues tool handler
	s.AddTool(listIssuesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received list_issues tool request", "args", args)

		// Extract project_id
		projectIDFloat, ok := args["project_id"].(float64)
		if !ok {
			debugLogger.Error("project_id is not a number", "value", args["project_id"])
			return mcp.NewToolResultError("project_id must be a number"), nil
		}
		projectID := int(projectIDFloat)

		// Extract optional parameters
		opts := &app.ListIssuesOptions{
			State: "opened", // default
			Limit: 100,      // default
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

		debugLogger.Debug("Processing list_issues request", "project_id", projectID, "opts", opts)

		// Call the app method
		issues, err := appInstance.ListProjectIssues(projectID, opts)
		if err != nil {
			debugLogger.Error("Failed to list project issues", "error", err, "project_id", projectID)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list project issues: %v", err)), nil
		}

		// Convert issues to JSON
		jsonData, err := json.Marshal(issues)
		if err != nil {
			debugLogger.Error("Failed to marshal issues to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format issues response"), nil
		}

		debugLogger.Info("Successfully retrieved project issues", "count", len(issues), "project_id", projectID)
		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// Create create_issues tool
	createIssueTool := mcp.NewTool("create_issues",
		mcp.WithDescription("Create a new issue for a GitLab project by project ID"),
		mcp.WithNumber("project_id",
			mcp.Required(),
			mcp.Description("GitLab project ID"),
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

	// Add create_issues tool handler
	s.AddTool(createIssueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received create_issues tool request", "args", args)

		// Extract project_id
		projectIDFloat, ok := args["project_id"].(float64)
		if !ok {
			debugLogger.Error("project_id is not a number", "value", args["project_id"])
			return mcp.NewToolResultError("project_id must be a number"), nil
		}
		projectID := int(projectIDFloat)

		// Extract title (required)
		title, ok := args["title"].(string)
		if !ok || title == "" {
			debugLogger.Error("title is missing or not a string", "value", args["title"])
			return mcp.NewToolResultError("title must be a non-empty string"), nil
		}

		// Create options
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

		debugLogger.Debug("Processing create_issues request", "project_id", projectID, "title", title)

		// Call the app method
		issue, err := appInstance.CreateProjectIssue(projectID, opts)
		if err != nil {
			debugLogger.Error("Failed to create issue", "error", err, "project_id", projectID, "title", title)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create issue: %v", err)), nil
		}

		// Convert issue to JSON
		jsonData, err := json.Marshal(issue)
		if err != nil {
			debugLogger.Error("Failed to marshal issue to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format issue response"), nil
		}

		debugLogger.Info("Successfully created issue", "id", issue.ID, "iid", issue.IID, "project_id", projectID, "title", issue.Title)
		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
