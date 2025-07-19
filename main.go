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

	// Create list_issues tool
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

	// Add list_issues tool handler
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

	// Create create_issues tool
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

	// Add create_issues tool handler
	s.AddTool(createIssueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		debugLogger.Info("Successfully created issue", "id", issue.ID, "iid", issue.IID, "project_path", projectPath, "title", issue.Title)
		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// Create list_labels tool
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

	// Create update_issues tool
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

	// Add update_issues tool handler
	s.AddTool(updateIssueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		// Create options
		opts := &app.UpdateIssueOptions{}

		// Extract optional title
		if title, ok := args["title"].(string); ok && title != "" {
			opts.Title = title
		}

		// Extract optional description
		if description, ok := args["description"].(string); ok {
			opts.Description = description
		}

		// Extract optional state
		if state, ok := args["state"].(string); ok && state != "" {
			if state != "opened" && state != "closed" {
				debugLogger.Error("invalid state value", "state", state)
				return mcp.NewToolResultError("state must be 'opened' or 'closed'"), nil
			}
			opts.State = state
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
	})

	// Add list_labels tool handler
	s.AddTool(listLabelsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received list_labels tool request", "args", args)

		// Extract project_path
		projectPath, ok := args["project_path"].(string)
		if !ok || projectPath == "" {
			debugLogger.Error("project_path is not a valid string", "value", args["project_path"])
			return mcp.NewToolResultError("project_path must be a non-empty string"), nil
		}

		// Extract optional parameters
		opts := &app.ListLabelsOptions{
			WithCounts:            false, // default
			IncludeAncestorGroups: false, // default
			Limit:                 100,   // default
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
	})

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
