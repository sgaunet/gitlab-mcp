package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sgaunet/gitlab-mcp/internal/app"
)

// marshalToolResult converts a value to a JSON tool result, reporting a formatting
// failure as a tool error rather than a protocol error.
func marshalToolResult(value any, debugLogger *slog.Logger, subject string) *mcp.CallToolResult {
	jsonData, err := json.Marshal(value)
	if err != nil {
		debugLogger.Error("Failed to marshal response to JSON", "error", err, "subject", subject)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format %s response", subject))
	}
	return mcp.NewToolResultText(string(jsonData))
}

// parseListGroupProjectsOptions extracts list group projects options from arguments.
func parseListGroupProjectsOptions(args map[string]any) *app.ListGroupProjectsOptions {
	opts := &app.ListGroupProjectsOptions{
		IncludeSubgroups: true, // default: recurse through all descendant subgroups
		Limit:            defaultLimit,
	}

	if includeSubgroups, ok := args["include_subgroups"].(bool); ok {
		opts.IncludeSubgroups = includeSubgroups
	}
	if includeArchived, ok := args["include_archived"].(bool); ok {
		opts.IncludeArchived = includeArchived
	}
	if search, ok := args["search"].(string); ok && search != "" {
		opts.Search = search
	}
	if topic, ok := args["topic"].(string); ok && topic != "" {
		opts.Topic = topic
	}
	if limitFloat, ok := args["limit"].(float64); ok {
		opts.Limit = int64(limitFloat)
	}

	return opts
}

// setupListGroupProjectsTool creates and registers the list_group_projects tool.
func setupListGroupProjectsTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
	listGroupProjectsTool := mcp.NewTool("list_group_projects",
		mcp.WithDescription("List all projects of a GitLab group, recursively including every subgroup at any depth. "+
			"Returns each project's id, name, full namespaced path, description, topics, web URL and archived flag. "+
			"Use the returned path with the other tools of this server (list_issues, get_project_description, ...). "+
			"Set include_subgroups=false to list only projects owned directly by the group."),
		mcp.WithString("group_path",
			mcp.Required(),
			mcp.Description("GitLab group path including all parent namespaces (e.g., 'mygroup' or "+
				"'company/department/team'). Accepts nested group paths"),
		),
		mcp.WithBoolean("include_subgroups",
			mcp.Description("Recurse into all descendant subgroups. Defaults to true for comprehensive results. "+
				"Set to false to list only projects directly owned by the group"),
		),
		mcp.WithBoolean("include_archived",
			mcp.Description("Include archived projects in the results (default: false)"),
		),
		mcp.WithString("search",
			mcp.Description("Filter projects by name or path substring"),
		),
		mcp.WithString("topic",
			mcp.Description("Filter projects by topic/tag"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of projects to return (default: 100, max: 1000). "+
				"Results spanning several API pages are fetched automatically"),
		),
	)

	s.AddTool(listGroupProjectsTool, func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		debugLogger.Debug("Received list_group_projects tool request", "args", args)

		// Extract group_path
		groupPath, ok := args["group_path"].(string)
		if !ok || groupPath == "" {
			debugLogger.Error("group_path is not a valid string", "value", args["group_path"])
			return mcp.NewToolResultError("group_path must be a non-empty string"), nil
		}

		opts := parseListGroupProjectsOptions(args)

		debugLogger.Debug("Processing list_group_projects request", "group_path", groupPath, "opts", opts)

		// Call the app method
		projects, err := appInstance.ListGroupProjects(groupPath, opts)
		if err != nil {
			debugLogger.Error("Failed to list group projects", "error", err, "group_path", groupPath)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list group projects: %v", err)), nil
		}

		debugLogger.Info("Successfully retrieved group projects", "count", len(projects), "group_path", groupPath)
		return marshalToolResult(projects, debugLogger, "projects"), nil
	})
}
