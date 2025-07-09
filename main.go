package main

import (
	"context"
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

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
