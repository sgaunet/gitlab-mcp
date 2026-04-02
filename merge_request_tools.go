package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sgaunet/gitlab-mcp/internal/app"
)

var (
	errProjectPathRequired  = errors.New("project_path must be a non-empty string")
	errTitleRequired        = errors.New("title must be a non-empty string")
	errSourceBranchRequired = errors.New("source_branch must be a non-empty string")
	errTargetBranchRequired = errors.New("target_branch must be a non-empty string")
)

// parseLabels extracts labels from interface array.
func parseLabels(labelsInterface any) []string {
	labels := make([]string, 0)
	if arr, ok := labelsInterface.([]any); ok {
		for _, label := range arr {
			if labelStr, ok := label.(string); ok {
				labels = append(labels, labelStr)
			}
		}
	}
	return labels
}

// parseIDArray extracts int64 IDs from interface array.
func parseIDArray(idsInterface any) []int64 {
	ids := make([]int64, 0)
	if arr, ok := idsInterface.([]any); ok {
		for _, id := range arr {
			if idFloat, ok := id.(float64); ok {
				ids = append(ids, int64(idFloat))
			}
		}
	}
	return ids
}

// parseListMROptions extracts list merge request options from arguments.
func parseListMROptions(args map[string]any) *app.ListMergeRequestsOptions {
	opts := &app.ListMergeRequestsOptions{State: "opened", Limit: defaultLimit}

	if state, ok := args["state"].(string); ok && state != "" {
		opts.State = state
	}
	if labelsInterface, ok := args["labels"]; ok {
		opts.Labels = parseLabels(labelsInterface)
	}
	if author, ok := args["author"].(string); ok && author != "" {
		opts.Author = author
	}
	if search, ok := args["search"].(string); ok && search != "" {
		opts.Search = search
	}
	if limitFloat, ok := args["limit"].(float64); ok {
		opts.Limit = int64(limitFloat)
	}

	return opts
}

// parseUpdateMROptions extracts update merge request options from arguments.
func parseUpdateMROptions(args map[string]any) *app.UpdateMergeRequestOptions {
	opts := &app.UpdateMergeRequestOptions{}

	if title, ok := args["title"].(string); ok && title != "" {
		opts.Title = title
	}
	if desc, ok := args["description"].(string); ok {
		opts.Description = desc
	}
	if state, ok := args["state"].(string); ok && state != "" {
		opts.State = state
	}
	if labelsInterface, ok := args["labels"]; ok {
		opts.Labels = parseLabels(labelsInterface)
	}
	if assigneeIDs, ok := args["assignee_ids"]; ok {
		opts.AssigneeIDs = parseIDArray(assigneeIDs)
	}
	if reviewerIDs, ok := args["reviewer_ids"]; ok {
		opts.ReviewerIDs = parseIDArray(reviewerIDs)
	}

	return opts
}

// setupListMergeRequestsTool creates and registers the list_merge_requests tool.
func setupListMergeRequestsTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	tool := mcp.NewTool("list_merge_requests",
		mcp.WithDescription("List merge requests for a GitLab project by project path with filtering options"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("GitLab project path including all namespaces"),
		),
		mcp.WithString("state",
			mcp.Description("Filter by state: opened, closed, locked, merged (default: opened)"),
		),
		mcp.WithArray("labels",
			mcp.Description("Array of labels to filter by"),
		),
		mcp.WithString("author",
			mcp.Description("Filter by author username"),
		),
		mcp.WithString("search",
			mcp.Description("Search in title and description"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results (default: 100, max: 100)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received list_merge_requests tool request", "args", args)

		projectPath, ok := args["project_path"].(string)
		if !ok || projectPath == "" {
			return mcp.NewToolResultError("project_path must be a non-empty string"), nil
		}

		opts := parseListMROptions(args)
		mrs, err := appInstance.ListProjectMergeRequests(projectPath, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list merge requests: %v", err)), nil
		}

		jsonData, err := json.Marshal(mrs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal merge requests: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})
}

// extractRequiredString validates and extracts a required string argument.
func extractRequiredString(args map[string]any, key string, errVal error) (string, error) {
	val, ok := args[key].(string)
	if !ok || val == "" {
		return "", errVal
	}
	return val, nil
}

// extractCreateMRArgs validates and extracts arguments for creating a merge request.
func extractCreateMRArgs(args map[string]any) (string, *app.CreateMergeRequestOptions, error) {
	projectPath, err := extractRequiredString(args, "project_path", errProjectPathRequired)
	if err != nil {
		return "", nil, err
	}

	title, err := extractRequiredString(args, "title", errTitleRequired)
	if err != nil {
		return "", nil, err
	}

	sourceBranch, err := extractRequiredString(args, "source_branch", errSourceBranchRequired)
	if err != nil {
		return "", nil, err
	}

	targetBranch, err := extractRequiredString(args, "target_branch", errTargetBranchRequired)
	if err != nil {
		return "", nil, err
	}

	opts := &app.CreateMergeRequestOptions{
		Title:        title,
		SourceBranch: sourceBranch,
		TargetBranch: targetBranch,
	}

	if desc, ok := args["description"].(string); ok {
		opts.Description = desc
	}
	if labelsInterface, ok := args["labels"]; ok {
		opts.Labels = parseLabels(labelsInterface)
	}
	if assigneeIDs, ok := args["assignee_ids"]; ok {
		opts.AssigneeIDs = parseIDArray(assigneeIDs)
	}
	if reviewerIDs, ok := args["reviewer_ids"]; ok {
		opts.ReviewerIDs = parseIDArray(reviewerIDs)
	}

	return projectPath, opts, nil
}

// setupCreateMergeRequestTool creates and registers the create_merge_request tool.
func setupCreateMergeRequestTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	tool := mcp.NewTool("create_merge_request",
		mcp.WithDescription("Create a new merge request in a GitLab project"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("GitLab project path")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Merge request title")),
		mcp.WithString("source_branch", mcp.Required(), mcp.Description("Source branch name")),
		mcp.WithString("target_branch", mcp.Required(), mcp.Description("Target branch name")),
		mcp.WithString("description", mcp.Description("Merge request description")),
		mcp.WithArray("labels", mcp.Description("Array of labels")),
		mcp.WithArray("assignee_ids", mcp.Description("Array of assignee user IDs")),
		mcp.WithArray("reviewer_ids", mcp.Description("Array of reviewer user IDs")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received create_merge_request tool request", "args", args)

		projectPath, opts, err := extractCreateMRArgs(args)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		mr, err := appInstance.CreateProjectMergeRequest(projectPath, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create merge request: %v", err)), nil
		}

		jsonData, err := json.Marshal(mr)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal merge request: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})
}

// setupGetMergeRequestTool creates and registers the get_merge_request tool.
func setupGetMergeRequestTool(s *server.MCPServer, appInstance *app.App, _ *slog.Logger) {
	tool := mcp.NewTool("get_merge_request",
		mcp.WithDescription("Get details of a specific merge request"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("GitLab project path")),
		mcp.WithNumber("mr_iid", mcp.Required(), mcp.Description("Merge request internal ID (IID)")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		projectPath, _ := args["project_path"].(string)
		mrIIDFloat, _ := args["mr_iid"].(float64)
		mrIID := int64(mrIIDFloat)

		mr, err := appInstance.GetMergeRequest(projectPath, mrIID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get merge request: %v", err)), nil
		}

		jsonData, err := json.Marshal(mr)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal merge request: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})
}

// setupUpdateMergeRequestTool creates and registers the update_merge_request tool.
func setupUpdateMergeRequestTool(s *server.MCPServer, appInstance *app.App, _ *slog.Logger) {
	tool := mcp.NewTool("update_merge_request",
		mcp.WithDescription("Update an existing merge request"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("GitLab project path")),
		mcp.WithNumber("mr_iid", mcp.Required(), mcp.Description("Merge request IID to update")),
		mcp.WithString("title", mcp.Description("Updated title")),
		mcp.WithString("description", mcp.Description("Updated description")),
		mcp.WithString("state", mcp.Description("State: opened, closed")),
		mcp.WithArray("labels", mcp.Description("Array of labels")),
		mcp.WithArray("assignee_ids", mcp.Description("Array of assignee user IDs")),
		mcp.WithArray("reviewer_ids", mcp.Description("Array of reviewer user IDs")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		projectPath, _ := args["project_path"].(string)
		mrIIDFloat, _ := args["mr_iid"].(float64)
		mrIID := int64(mrIIDFloat)

		opts := parseUpdateMROptions(args)
		mr, err := appInstance.UpdateMergeRequest(projectPath, mrIID, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update merge request: %v", err)), nil
		}

		jsonData, err := json.Marshal(mr)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal merge request: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})
}

// setupMergeMergeRequestTool creates and registers the merge_merge_request tool.
func setupMergeMergeRequestTool(s *server.MCPServer, appInstance *app.App, _ *slog.Logger) {
	tool := mcp.NewTool("merge_merge_request",
		mcp.WithDescription("Merge an approved merge request"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("GitLab project path")),
		mcp.WithNumber("mr_iid", mcp.Required(), mcp.Description("Merge request IID to merge")),
		mcp.WithString("merge_commit_message", mcp.Description("Custom merge commit message")),
		mcp.WithString("squash_commit_message", mcp.Description("Custom squash commit message")),
		mcp.WithBoolean("squash", mcp.Description("Squash commits before merging")),
		mcp.WithBoolean("should_remove_source_branch", mcp.Description("Remove source branch after merge")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		projectPath, _ := args["project_path"].(string)
		mrIIDFloat, _ := args["mr_iid"].(float64)
		mrIID := int64(mrIIDFloat)

		opts := &app.MergeMergeRequestOptions{}
		if msg, ok := args["merge_commit_message"].(string); ok && msg != "" {
			opts.MergeCommitMessage = msg
		}
		if msg, ok := args["squash_commit_message"].(string); ok && msg != "" {
			opts.SquashCommitMessage = msg
		}
		if squash, ok := args["squash"].(bool); ok {
			opts.Squash = squash
		}
		if remove, ok := args["should_remove_source_branch"].(bool); ok {
			opts.ShouldRemoveSourceBranch = remove
		}

		mr, err := appInstance.MergeMergeRequest(projectPath, mrIID, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to merge request: %v", err)), nil
		}

		jsonData, err := json.Marshal(mr)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal merge request: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})
}

// setupGetMergeRequestDiffTool creates and registers the get_merge_request_diff tool.
func setupGetMergeRequestDiffTool(s *server.MCPServer, appInstance *app.App, _ *slog.Logger) {
	tool := mcp.NewTool("get_merge_request_diff",
		mcp.WithDescription("Get the diff for a merge request"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("GitLab project path")),
		mcp.WithNumber("mr_iid", mcp.Required(), mcp.Description("Merge request IID")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		projectPath, _ := args["project_path"].(string)
		mrIIDFloat, _ := args["mr_iid"].(float64)
		mrIID := int64(mrIIDFloat)

		diff, err := appInstance.GetMergeRequestDiff(projectPath, mrIID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get merge request diff: %v", err)), nil
		}

		return mcp.NewToolResultText(diff), nil
	})
}

// setupApproveMergeRequestTool creates and registers the approve_merge_request tool.
func setupApproveMergeRequestTool(s *server.MCPServer, appInstance *app.App, _ *slog.Logger) {
	tool := mcp.NewTool("approve_merge_request",
		mcp.WithDescription("Approve a merge request"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("GitLab project path")),
		mcp.WithNumber("mr_iid", mcp.Required(), mcp.Description("Merge request IID to approve")),
		mcp.WithString("sha", mcp.Description("Optional: specific commit SHA to approve")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		projectPath, _ := args["project_path"].(string)
		mrIIDFloat, _ := args["mr_iid"].(float64)
		mrIID := int64(mrIIDFloat)

		opts := &app.ApproveMergeRequestOptions{}
		if sha, ok := args["sha"].(string); ok && sha != "" {
			opts.SHA = sha
		}

		result, err := appInstance.ApproveMergeRequest(projectPath, mrIID, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to approve merge request: %v", err)), nil
		}

		return mcp.NewToolResultText(result), nil
	})
}

// setupAddMergeRequestNoteTool creates and registers the add_merge_request_note tool.
func setupAddMergeRequestNoteTool(s *server.MCPServer, appInstance *app.App, _ *slog.Logger) {
	tool := mcp.NewTool("add_merge_request_note",
		mcp.WithDescription("Add a comment/note to a merge request"),
		mcp.WithString("project_path", mcp.Required(), mcp.Description("GitLab project path")),
		mcp.WithNumber("mr_iid", mcp.Required(), mcp.Description("Merge request IID")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Comment body text")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		projectPath, _ := args["project_path"].(string)
		mrIIDFloat, _ := args["mr_iid"].(float64)
		mrIID := int64(mrIIDFloat)
		body, _ := args["body"].(string)

		opts := &app.AddMergeRequestNoteOptions{Body: body}
		note, err := appInstance.AddMergeRequestNote(projectPath, mrIID, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to add merge request note: %v", err)), nil
		}

		jsonData, err := json.Marshal(note)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal note: %w", err)
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})
}
