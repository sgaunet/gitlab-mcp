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
	ErrInvalidStateValue            = errors.New("state must be 'opened' or 'closed'")
	ErrGroupPathRequired            = errors.New("group_path is required and must be a non-empty string")
	ErrEpicIIDRequired              = errors.New("epic_iid is required and must be a number")
	ErrEpicIIDMustBePositive        = errors.New("epic_iid must be greater than 0")
	ErrProjectPathRequired          = errors.New("project_path is required and must be a non-empty string")
	ErrIssueIIDRequired             = errors.New("issue_iid is required and must be a number")
	ErrIssueIIDMustBePositive       = errors.New("issue_iid must be greater than 0")
)

// setupListIssuesTool creates and registers the list_issues tool.
func setupListIssuesTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	listIssuesTool := mcp.NewTool("list_issues",
		mcp.WithDescription("List issues for a GitLab project by project path"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("GitLab project path including all namespaces (e.g., 'namespace/project-name' or "+
				"'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL"),
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
			State: "opened",     // default
			Limit: defaultLimit, // default
		}

		if state, ok := args["state"].(string); ok && state != "" {
			opts.State = state
		}

		if labels, ok := args["labels"].(string); ok && labels != "" {
			opts.Labels = labels
		}

		if limitFloat, ok := args["limit"].(float64); ok {
			opts.Limit = int64(limitFloat)
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
			mcp.Description("GitLab project path including all namespaces (e.g., 'namespace/project-name' or "+
				"'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Issue title"),
		),
		mcp.WithString("description",
			mcp.Description("Issue description"),
		),
		mcp.WithArray("labels",
			mcp.Description("Array of labels to assign to the issue. Labels must exist in the project. "+
				"Use list_labels tool to see available labels. Set GITLAB_VALIDATE_LABELS=false to disable validation."),
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
func extractCreateIssueOptions(args map[string]any, title string) *app.CreateIssueOptions {
	opts := &app.CreateIssueOptions{
		Title: title,
	}

	// Extract optional description
	if description, ok := args["description"].(string); ok {
		opts.Description = description
	}

	// Extract optional labels
	if labelsInterface, ok := args["labels"].([]any); ok {
		labels := make([]string, 0, len(labelsInterface))
		for _, label := range labelsInterface {
			if labelStr, ok := label.(string); ok {
				labels = append(labels, labelStr)
			}
		}
		opts.Labels = labels
	}

	// Extract optional assignees
	if assigneesInterface, ok := args["assignees"].([]any); ok {
		assignees := make([]int64, 0, len(assigneesInterface))
		for _, assignee := range assigneesInterface {
			if assigneeFloat, ok := assignee.(float64); ok {
				assignees = append(assignees, int64(assigneeFloat))
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
			mcp.Description("GitLab project path including all namespaces (e.g., 'namespace/project-name' or "+
				"'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL"),
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
		issueIID := int64(issueIIDFloat)

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
func extractUpdateIssueOptions(args map[string]any, debugLogger *slog.Logger) (*app.UpdateIssueOptions, error) {
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
func extractUpdateStringFields(args map[string]any, opts *app.UpdateIssueOptions) {
	if title, ok := args["title"].(string); ok && title != "" {
		opts.Title = title
	}

	if description, ok := args["description"].(string); ok {
		opts.Description = description
	}
}

// extractUpdateState extracts and validates the state field.
func extractUpdateState(args map[string]any, opts *app.UpdateIssueOptions, debugLogger *slog.Logger) error {
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
func extractUpdateArrayFields(args map[string]any, opts *app.UpdateIssueOptions) {
	// Extract optional labels
	if labelsInterface, ok := args["labels"].([]any); ok {
		labels := make([]string, 0, len(labelsInterface))
		for _, label := range labelsInterface {
			if labelStr, ok := label.(string); ok {
				labels = append(labels, labelStr)
			}
		}
		opts.Labels = labels
	}

	// Extract optional assignees
	if assigneesInterface, ok := args["assignees"].([]any); ok {
		assignees := make([]int64, 0, len(assigneesInterface))
		for _, assignee := range assigneesInterface {
			if assigneeFloat, ok := assignee.(float64); ok {
				assignees = append(assignees, int64(assigneeFloat))
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
			mcp.Description("GitLab project path including all namespaces (e.g., 'namespace/project-name' or "+
				"'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL"),
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
func extractListLabelsOptions(args map[string]any) *app.ListLabelsOptions {
	opts := &app.ListLabelsOptions{
		WithCounts:            false,        // default
		IncludeAncestorGroups: false,        // default
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
		opts.Limit = int64(limitFloat)
	}

	return opts
}

// setupListEpicsTool creates and registers the list_epics tool.
func setupListEpicsTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	listEpicsTool := mcp.NewTool("list_epics",
		mcp.WithDescription("List epics for a GitLab group by group path. "+
			"Note: Epics require GitLab Premium or Ultimate tier. "+
			"Free/Starter tier instances will return a helpful error message."),
		mcp.WithString("group_path",
			mcp.Required(),
			mcp.Description("GitLab group path (e.g., 'myorg' or 'parent/subgroup'). "+
				"Groups contain multiple projects and use epics to organize work across projects."),
		),
		mcp.WithString("state",
			mcp.Description("Filter by epic state: opened, closed, or all (default: opened)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of epics to return (default: 100, max: 100)"),
		),
	)

	s.AddTool(listEpicsTool, handleListEpicsRequest(appInstance, debugLogger))
}

// handleListEpicsRequest handles the list_epics tool request.
func handleListEpicsRequest(
	appInstance *app.App,
	debugLogger *slog.Logger,
) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received list_epics tool request", "args", args)

		// Extract group_path
		groupPath, ok := args["group_path"].(string)
		if !ok || groupPath == "" {
			debugLogger.Error("group_path is not a valid string", "value", args["group_path"])
			return mcp.NewToolResultError("group_path must be a non-empty string"), nil
		}

		// Extract optional parameters
		opts := &app.ListEpicsOptions{
			State: "opened",
			Limit: defaultLimit,
		}

		if state, ok := args["state"].(string); ok && state != "" {
			opts.State = state
		}
		if limitFloat, ok := args["limit"].(float64); ok {
			opts.Limit = int64(limitFloat)
		}

		debugLogger.Debug("Processing list_epics request", "group_path", groupPath, "opts", opts)

		// Call the app method
		epics, err := appInstance.ListGroupEpics(groupPath, opts)
		if err != nil {
			debugLogger.Error("Failed to list group epics", "error", err, "group_path", groupPath)

			if errors.Is(err, app.ErrEpicsTierRequired) {
				return mcp.NewToolResultError(fmt.Sprintf(
					"Failed to list epics: %v\n\n"+
						"Epics are a GitLab Premium/Ultimate feature. If you're on a Free tier, "+
						"consider using issues for work tracking instead. "+
						"See: https://docs.gitlab.com/ee/user/group/epics/",
					err,
				)), nil
			}

			return mcp.NewToolResultError(fmt.Sprintf("Failed to list group epics: %v", err)), nil
		}

		// Convert to JSON
		jsonData, err := json.Marshal(epics)
		if err != nil {
			debugLogger.Error("Failed to marshal epics to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format epics response"), nil
		}

		debugLogger.Info("Successfully retrieved group epics", "count", len(epics), "group_path", groupPath)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// setupCreateEpicTool creates and registers the create_epic tool.
func setupCreateEpicTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	createEpicTool := mcp.NewTool("create_epic",
		mcp.WithDescription("Create a new epic in a GitLab group. "+
			"Note: Epics require GitLab Premium or Ultimate tier."),
		mcp.WithString("group_path",
			mcp.Required(),
			mcp.Description("GitLab group path (e.g., 'myorg' or 'parent/subgroup')"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Epic title"),
		),
		mcp.WithString("description",
			mcp.Description("Epic description (optional)"),
		),
		mcp.WithArray("labels",
			mcp.Description("Array of label names (optional)"),
		),
		mcp.WithString("start_date",
			mcp.Description("Start date in YYYY-MM-DD format (optional)"),
		),
		mcp.WithString("due_date",
			mcp.Description("Due date in YYYY-MM-DD format (optional)"),
		),
		mcp.WithBoolean("confidential",
			mcp.Description("Whether epic is confidential (default: false)"),
		),
	)

	s.AddTool(createEpicTool, handleCreateEpicRequest(appInstance, debugLogger))
}

// handleCreateEpicRequest handles the create_epic tool request.
func handleCreateEpicRequest(
	appInstance *app.App,
	debugLogger *slog.Logger,
) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received create_epic tool request", "args", args)

		// Extract and validate parameters
		groupPath, opts, validationErr := extractCreateEpicParams(args, debugLogger)
		if validationErr != nil {
			return validationErr, nil
		}

		debugLogger.Debug("Processing create_epic request", "group_path", groupPath, "opts", opts)

		// Call the app method
		epic, appErr := appInstance.CreateGroupEpic(groupPath, opts)
		if appErr != nil {
			return handleCreateEpicError(appErr, groupPath, debugLogger), nil
		}

		// Convert to JSON
		jsonData, err := json.Marshal(epic)
		if err != nil {
			debugLogger.Error("Failed to marshal epic to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format epic response"), nil
		}

		debugLogger.Info("Successfully created epic", "id", epic.ID, "iid", epic.IID, "group_path", groupPath)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// setupAddIssueToEpicTool creates and registers the add_issue_to_epic tool.
func setupAddIssueToEpicTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	addIssueToEpicTool := mcp.NewTool("add_issue_to_epic",
		mcp.WithDescription("Attach an issue to an epic (Premium/Ultimate tier). "+
			"Note: The Epics API is deprecated and will be removed in GitLab API v5. "+
			"Consider using Work Items API for future implementations."),
		mcp.WithString("group_path",
			mcp.Required(),
			mcp.Description("GitLab group path containing the epic (e.g., 'myorg' or 'parent/subgroup'). "+
				"Run 'git remote -v' to find the full path from the repository URL"),
		),
		mcp.WithNumber("epic_iid",
			mcp.Required(),
			mcp.Description("Epic internal ID (IID) to attach the issue to"),
		),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("GitLab project path containing the issue (e.g., 'namespace/project-name')"),
		),
		mcp.WithNumber("issue_iid",
			mcp.Required(),
			mcp.Description("Issue internal ID (IID) to attach to the epic"),
		),
	)

	s.AddTool(addIssueToEpicTool, handleAddIssueToEpicRequest(appInstance, debugLogger))
}

// handleAddIssueToEpicRequest handles the add_issue_to_epic tool request.
func handleAddIssueToEpicRequest(
	appInstance *app.App,
	debugLogger *slog.Logger,
) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received add_issue_to_epic tool request", "args", args)

		// Extract and validate parameters
		opts, err := extractAddIssueToEpicParams(args)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		debugLogger.Debug("Processing add_issue_to_epic request", "opts", opts)

		// Call the app method
		epicIssue, err := appInstance.AddIssueToEpic(opts)
		if err != nil {
			if errors.Is(err, app.ErrEpicsTierRequired) {
				return mcp.NewToolResultError(
					"This feature requires GitLab Premium or Ultimate. " +
						"See https://docs.gitlab.com/ee/user/group/epics/ for more information.",
				), nil
			}
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Convert to JSON
		jsonData, err := json.Marshal(epicIssue)
		if err != nil {
			debugLogger.Error("Failed to marshal epic issue to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format epic issue response"), nil
		}

		debugLogger.Info("Successfully added issue to epic",
			"issue_id", epicIssue.ID, "issue_iid", epicIssue.IID, "epic_iid", epicIssue.EpicIID)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// setupAddIssueNoteTool creates and registers the add_issue_note tool.
func setupAddIssueNoteTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	addIssueNoteTool := mcp.NewTool("add_issue_note",
		mcp.WithDescription("Add a note/comment to an existing issue for a GitLab project by project path"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("GitLab project path including all namespaces (e.g., 'namespace/project-name' or "+
				"'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL"),
		),
		mcp.WithNumber("issue_iid",
			mcp.Required(),
			mcp.Description("Issue internal ID (IID) to add note to"),
		),
		mcp.WithString("body",
			mcp.Required(),
			mcp.Description("Note/comment body text"),
		),
	)

	s.AddTool(addIssueNoteTool, handleAddIssueNoteRequest(appInstance, debugLogger))
}

// handleNoteRequest handles the add_issue_note tool request.
func handleNoteRequest(
	appInstance *app.App,
	debugLogger *slog.Logger,
) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received add_issue_note tool request", "args", args)

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
		issueIID := int64(issueIIDFloat)

		// Extract body (required)
		body, ok := args["body"].(string)
		if !ok || body == "" {
			debugLogger.Error("body is missing or not a string", "value", args["body"])
			return mcp.NewToolResultError("body must be a non-empty string"), nil
		}

		// Create options
		opts := &app.AddIssueNoteOptions{
			Body: body,
		}

		debugLogger.Debug("Processing add_issue_note request", "project_path", projectPath, "issue_iid", issueIID)

		// Call the app method
		note, err := appInstance.AddIssueNote(projectPath, issueIID, opts)
		if err != nil {
			debugLogger.Error("Failed to add issue note", "error", err, "project_path", projectPath, "issue_iid", issueIID)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to add issue note: %v", err)), nil
		}

		// Convert note to JSON
		jsonData, err := json.Marshal(note)
		if err != nil {
			debugLogger.Error("Failed to marshal note to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format note response"), nil
		}

		debugLogger.Info("Successfully added note to issue",
			"note_id", note.ID,
			"project_path", projectPath,
			"issue_iid", issueIID)
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

var handleAddIssueNoteRequest = handleNoteRequest


// setupProjectInfoTool creates a generic project info tool handler.
func setupProjectInfoTool(
	s *server.MCPServer,
	debugLogger *slog.Logger,
	toolName, toolDescription, actionLog, errorLog, successLog string,
	handler func(string) (*app.ProjectInfo, error),
) {
	tool := mcp.NewTool(toolName,
		mcp.WithDescription(toolDescription),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("GitLab project path including all namespaces (e.g., 'namespace/project-name' or "+
				"'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received "+toolName+" tool request", "args", args)

		// Extract project_path
		projectPath, ok := args["project_path"].(string)
		if !ok || projectPath == "" {
			debugLogger.Error("project_path is not a valid string", "value", args["project_path"])
			return mcp.NewToolResultError("project_path must be a non-empty string"), nil
		}

		debugLogger.Debug("Processing "+actionLog+" request", "project_path", projectPath)

		// Call the app method
		projectInfo, err := handler(projectPath)
		if err != nil {
			debugLogger.Error(errorLog, "error", err, "project_path", projectPath)
			return mcp.NewToolResultError(fmt.Sprintf(errorLog+": %v", err)), nil
		}

		// Convert to JSON
		jsonData, err := json.Marshal(projectInfo)
		if err != nil {
			debugLogger.Error("Failed to marshal project info to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format project info response"), nil
		}

		debugLogger.Info(successLog, "project_path", projectPath)
		return mcp.NewToolResultText(string(jsonData)), nil
	})
}

// setupGetProjectDescriptionTool creates and registers the get_project_description tool.
func setupGetProjectDescriptionTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	setupProjectInfoTool(
		s, debugLogger,
		"get_project_description",
		"Get the description of a GitLab project by project path",
		"get_project_description",
		"Failed to get project description",
		"Successfully retrieved project description",
		appInstance.GetProjectDescription,
	)
}

// setupUpdateProjectDescriptionTool creates and registers the update_project_description tool.
func setupUpdateProjectDescriptionTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	updateProjectDescriptionTool := mcp.NewTool("update_project_description",
		mcp.WithDescription("Update the description of a GitLab project by project path"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("GitLab project path including all namespaces (e.g., 'namespace/project-name' or "+
				"'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL"),
		),
		mcp.WithString("description",
			mcp.Required(),
			mcp.Description("The new description for the project"),
		),
	)

	s.AddTool(updateProjectDescriptionTool, func(ctx context.Context, request mcp.CallToolRequest) (
		*mcp.CallToolResult, error,
	) {
		args := request.GetArguments()
		debugLogger.Debug("Received update_project_description tool request", "args", args)

		// Extract project_path
		projectPath, ok := args["project_path"].(string)
		if !ok || projectPath == "" {
			debugLogger.Error("project_path is not a valid string", "value", args["project_path"])
			return mcp.NewToolResultError("project_path must be a non-empty string"), nil
		}

		// Extract description
		description, ok := args["description"].(string)
		if !ok {
			debugLogger.Error("description is not a valid string", "value", args["description"])
			return mcp.NewToolResultError("description must be a string"), nil
		}

		debugLogger.Debug("Processing update_project_description request", "project_path", projectPath)

		// Call the app method
		projectInfo, err := appInstance.UpdateProjectDescription(projectPath, description)
		if err != nil {
			debugLogger.Error("Failed to update project description", "error", err, "project_path", projectPath)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update project description: %v", err)), nil
		}

		// Convert to JSON
		jsonData, err := json.Marshal(projectInfo)
		if err != nil {
			debugLogger.Error("Failed to marshal project info to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format project info response"), nil
		}

		debugLogger.Info("Successfully updated project description", "project_path", projectPath)
		return mcp.NewToolResultText(string(jsonData)), nil
	})
}

// setupGetProjectTopicsTool creates and registers the get_project_topics tool.
func setupGetProjectTopicsTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	setupProjectInfoTool(
		s, debugLogger,
		"get_project_topics",
		"Get the topics/tags of a GitLab project by project path",
		"get_project_topics",
		"Failed to get project topics",
		"Successfully retrieved project topics",
		appInstance.GetProjectTopics,
	)
}

// setupUpdateProjectTopicsTool creates and registers the update_project_topics tool.
func setupUpdateProjectTopicsTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	updateProjectTopicsTool := mcp.NewTool("update_project_topics",
		mcp.WithDescription("Update the topics/tags of a GitLab project (replaces all existing topics)"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("GitLab project path including all namespaces (e.g., 'namespace/project-name' or "+
				"'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL"),
		),
		mcp.WithArray("topics",
			mcp.Required(),
			mcp.Description("Array of topic strings to set for the project (replaces all existing topics)"),
		),
	)

	s.AddTool(updateProjectTopicsTool, func(ctx context.Context, request mcp.CallToolRequest) (
		*mcp.CallToolResult, error,
	) {
		args := request.GetArguments()
		debugLogger.Debug("Received update_project_topics tool request", "args", args)

		// Extract project_path
		projectPath, ok := args["project_path"].(string)
		if !ok || projectPath == "" {
			debugLogger.Error("project_path is not a valid string", "value", args["project_path"])
			return mcp.NewToolResultError("project_path must be a non-empty string"), nil
		}

		// Extract topics
		topicsInterface, ok := args["topics"].([]any)
		if !ok {
			debugLogger.Error("topics is not an array", "value", args["topics"])
			return mcp.NewToolResultError("topics must be an array of strings"), nil
		}

		// Convert interface array to string array
		topics := make([]string, 0, len(topicsInterface))
		for _, topic := range topicsInterface {
			if topicStr, ok := topic.(string); ok {
				topics = append(topics, topicStr)
			} else {
				debugLogger.Error("topic item is not a string", "value", topic)
				return mcp.NewToolResultError("all topics must be strings"), nil
			}
		}

		debugLogger.Debug("Processing update_project_topics request", "project_path", projectPath, "topics", topics)

		// Call the app method
		projectInfo, err := appInstance.UpdateProjectTopics(projectPath, topics)
		if err != nil {
			debugLogger.Error("Failed to update project topics", "error", err, "project_path", projectPath)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update project topics: %v", err)), nil
		}

		// Convert to JSON
		jsonData, err := json.Marshal(projectInfo)
		if err != nil {
			debugLogger.Error("Failed to marshal project info to JSON", "error", err)
			return mcp.NewToolResultError("Failed to format project info response"), nil
		}

		debugLogger.Info("Successfully updated project topics", "project_path", projectPath, "topics_count", len(topics))
		return mcp.NewToolResultText(string(jsonData)), nil
	})
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
    
    • list_issues              - List issues for a GitLab project
    • create_issues            - Create new issues with metadata
    • update_issues            - Update existing issues
    • list_labels              - List project labels with filtering
    • add_issue_note           - Add notes/comments to existing issues
    • add_merge_request_note   - Add notes/comments to existing merge requests
    • create_merge_request     - Create new merge requests
    • get_project_description  - Get the description of a project
    • update_project_description - Update the project description
    • get_project_topics       - Get the topics/tags of a project
    • update_project_topics    - Update the project topics (replaces all)
    • list_epics               - List epics for a GitLab group (Premium/Ultimate)

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

// handleCommandLineFlags processes command line arguments and exits if necessary.
func handleCommandLineFlags() {
	var (
		showHelp        = flag.Bool("h", false, "Show help message")
		showHelpLong    = flag.Bool("help", false, "Show help message")
		showVersion     = flag.Bool("v", false, "Show version information")
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
}

// initializeApp creates and configures the application instance.
func initializeApp() (*app.App, *slog.Logger) {
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

	return appInstance, debugLogger
}

// extractCreateEpicParams extracts and validates parameters for create_epic.
func extractCreateEpicParams(
	args map[string]any,
	debugLogger *slog.Logger,
) (string, *app.CreateEpicOptions, *mcp.CallToolResult) {
	// Extract required parameters
	groupPath, ok := args["group_path"].(string)
	if !ok || groupPath == "" {
		debugLogger.Error("group_path is not a valid string", "value", args["group_path"])
		return "", nil, mcp.NewToolResultError("group_path must be a non-empty string")
	}

	title, ok := args["title"].(string)
	if !ok || title == "" {
		debugLogger.Error("title is not a valid string", "value", args["title"])
		return "", nil, mcp.NewToolResultError("title must be a non-empty string")
	}

	// Build options with required and optional parameters
	opts := buildCreateEpicOptions(title, args)

	return groupPath, opts, nil
}

// extractAddIssueToEpicParams extracts and validates add_issue_to_epic parameters.
func extractAddIssueToEpicParams(args map[string]any) (*app.AddIssueToEpicOptions, error) {
	// Extract group_path
	groupPath, ok := args["group_path"].(string)
	if !ok || groupPath == "" {
		return nil, ErrGroupPathRequired
	}

	// Extract epic_iid
	epicIIDFloat, ok := args["epic_iid"].(float64)
	if !ok {
		return nil, ErrEpicIIDRequired
	}
	epicIID := int(epicIIDFloat)
	if epicIID <= 0 {
		return nil, ErrEpicIIDMustBePositive
	}

	// Extract project_path
	projectPath, ok := args["project_path"].(string)
	if !ok || projectPath == "" {
		return nil, ErrProjectPathRequired
	}

	// Extract issue_iid
	issueIIDFloat, ok := args["issue_iid"].(float64)
	if !ok {
		return nil, ErrIssueIIDRequired
	}
	issueIID := int64(issueIIDFloat)
	if issueIID <= 0 {
		return nil, ErrIssueIIDMustBePositive
	}

	return &app.AddIssueToEpicOptions{
		GroupPath:   groupPath,
		EpicIID:     epicIID,
		ProjectPath: projectPath,
		IssueIID:    issueIID,
	}, nil
}

// buildCreateEpicOptions builds CreateEpicOptions from request arguments.
func buildCreateEpicOptions(title string, args map[string]any) *app.CreateEpicOptions {
	opts := &app.CreateEpicOptions{
		Title: title,
	}

	// Extract optional parameters
	if desc, ok := args["description"].(string); ok {
		opts.Description = desc
	}

	if labelsRaw, ok := args["labels"].([]any); ok {
		opts.Labels = extractStringArray(labelsRaw)
	}

	if startDate, ok := args["start_date"].(string); ok && startDate != "" {
		opts.StartDate = startDate
	}

	if dueDate, ok := args["due_date"].(string); ok && dueDate != "" {
		opts.DueDate = dueDate
	}

	if conf, ok := args["confidential"].(bool); ok {
		opts.Confidential = conf
	}

	return opts
}

// handleCreateEpicError handles errors from CreateGroupEpic.
func handleCreateEpicError(err error, groupPath string, debugLogger *slog.Logger) *mcp.CallToolResult {
	debugLogger.Error("Failed to create epic", "error", err, "group_path", groupPath)

	if errors.Is(err, app.ErrEpicsTierRequired) {
		return mcp.NewToolResultError(fmt.Sprintf(
			"Failed to create epic: %v\n\n"+
				"Epics are a GitLab Premium/Ultimate feature. "+
				"See: https://docs.gitlab.com/ee/user/group/epics/",
			err,
		))
	}

	return mcp.NewToolResultError(fmt.Sprintf("Failed to create epic: %v", err))
}

// extractStringArray extracts a string array from an interface array.
func extractStringArray(raw []any) []string {
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

func registerAllTools(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	setupListIssuesTool(s, appInstance, debugLogger)
	setupCreateIssueTool(s, appInstance, debugLogger)
	setupUpdateIssueTool(s, appInstance, debugLogger)
	setupListLabelsTool(s, appInstance, debugLogger)
	setupAddIssueNoteTool(s, appInstance, debugLogger)
	setupGetProjectDescriptionTool(s, appInstance, debugLogger)
	setupUpdateProjectDescriptionTool(s, appInstance, debugLogger)
	setupGetProjectTopicsTool(s, appInstance, debugLogger)
	setupUpdateProjectTopicsTool(s, appInstance, debugLogger)
	setupListEpicsTool(s, appInstance, debugLogger)
	setupCreateEpicTool(s, appInstance, debugLogger)
	setupAddIssueToEpicTool(s, appInstance, debugLogger)
}

func main() {
	handleCommandLineFlags()
	appInstance, debugLogger := initializeApp()

	// Create MCP server
	s := server.NewMCPServer(
		"GitLab MCP Server",
		version,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)

	registerAllTools(s, appInstance, debugLogger)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
