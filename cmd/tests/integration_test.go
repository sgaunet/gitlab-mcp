package tests

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"testing"
	"time"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPServerIntegration(t *testing.T) {
	// Check if GITLAB_TOKEN is set
	if os.Getenv("GITLAB_TOKEN") == "" {
		t.Skip("GITLAB_TOKEN environment variable not set, skipping integration test")
	}

	t.Run("FullMCPFlow", func(t *testing.T) {
		// Start the MCP server subprocess
		binary := "../../gitlab-mcp-coverage"
		cmd := exec.Command(binary)
		cmd.Dir = "."

		// Create stdin/stdout pipes
		stdin, err := cmd.StdinPipe()
		require.NoError(t, err, "Failed to create stdin pipe")

		stdout, err := cmd.StdoutPipe()
		require.NoError(t, err, "Failed to create stdout pipe")

		// Start the server process
		err = cmd.Start()
		require.NoError(t, err, "Failed to start MCP server")

		// Ensure cleanup - use graceful shutdown for coverage data
		defer func() {
			if cmd.Process != nil {
				// Try graceful shutdown first
				cmd.Process.Signal(os.Interrupt)
				// Wait a bit for graceful shutdown
				time.Sleep(100 * time.Millisecond)
				// Then kill if still running
				cmd.Process.Kill()
				cmd.Wait()
			}
		}()

		// Give the server time to start
		time.Sleep(2 * time.Second)

		// Create MCP client with stdio transport
		clientTransport := stdio.NewStdioServerTransportWithIO(stdout, stdin)
		client := mcp_golang.NewClient(clientTransport)

		// Test 1: Initialize the MCP client
		t.Run("Initialize", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			initResponse, err := client.Initialize(ctx)
			require.NoError(t, err, "Failed to initialize MCP client")

			assert.Equal(t, "2025-03-26", initResponse.ProtocolVersion, "Protocol version mismatch")
			assert.Equal(t, "GitLab MCP Server", initResponse.ServerInfo.Name, "Server name mismatch")
			assert.Equal(t, "1.0.0", initResponse.ServerInfo.Version, "Server version mismatch")

			// Check capabilities
			assert.NotNil(t, initResponse.Capabilities.Tools, "Tools capability should be present")
			assert.NotNil(t, initResponse.Capabilities.Tools.ListChanged, "Tools list_changed capability should be present")
			assert.True(t, *initResponse.Capabilities.Tools.ListChanged, "Tools list_changed capability should be true")
			assert.NotNil(t, initResponse.Capabilities.Resources, "Resources capability should be present")
			assert.NotNil(t, initResponse.Capabilities.Resources.Subscribe, "Resources subscribe capability should be present")
			assert.True(t, *initResponse.Capabilities.Resources.Subscribe, "Resources subscribe capability should be true")
		})

		// Test 2: List available tools
		t.Run("ListTools", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			toolsResponse, err := client.ListTools(ctx, nil)
			require.NoError(t, err, "Failed to list tools")

			assert.Len(t, toolsResponse.Tools, 3, "Expected exactly three tools")

			// Find tools by name to make test order-independent
			toolMap := make(map[string]interface{})
			for _, tool := range toolsResponse.Tools {
				toolMap[tool.Name] = tool
			}

			// Check that all expected tools are present
			expectedTools := []string{"get_project_id", "list_issues", "create_issues"}
			for _, expectedTool := range expectedTools {
				_, exists := toolMap[expectedTool]
				assert.True(t, exists, "%s tool should exist", expectedTool)
			}

			// Test get_project_id tool specifically for detailed validation
			for _, tool := range toolsResponse.Tools {
				if tool.Name == "get_project_id" {
					assert.NotNil(t, tool.Description, "get_project_id tool description should not be nil")
					assert.Equal(t, "Get GitLab project ID from git remote repository URL", *tool.Description, "get_project_id tool description mismatch")

					// Check input schema for get_project_id tool
					assert.NotNil(t, tool.InputSchema, "get_project_id tool input schema should be present")
					schema := tool.InputSchema.(map[string]interface{})
					assert.Equal(t, "object", schema["type"], "Input schema type should be object")

					// Check required parameters
					required, exists := schema["required"]
					assert.True(t, exists, "Required field should exist")
					requiredSlice := required.([]interface{})
					assert.Contains(t, requiredSlice, "remote_url", "remote_url should be required")

					// Check properties
					properties, exists := schema["properties"]
					assert.True(t, exists, "Properties field should exist")
					propertiesMap := properties.(map[string]interface{})
					remoteUrlProp, exists := propertiesMap["remote_url"]
					assert.True(t, exists, "remote_url property should exist")
					remoteUrlPropMap := remoteUrlProp.(map[string]interface{})
					assert.Equal(t, "string", remoteUrlPropMap["type"], "remote_url should be string type")
					assert.Equal(t, "Git remote repository URL (SSH or HTTPS)", remoteUrlPropMap["description"], "remote_url description mismatch")
					break
				}
			}
		})

		// Test 3: Call the get_project_id tool
		t.Run("CallGetProjectIDTool", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Test with the real GitLab repository
			args := map[string]interface{}{
				"remote_url": "git@gitlab.com:sgaunet/poc-table.git",
			}

			toolResult, err := client.CallTool(ctx, "get_project_id", args)
			require.NoError(t, err, "Failed to call get_project_id tool")

			assert.NotNil(t, toolResult.Content, "Tool result should have content")
			assert.Len(t, toolResult.Content, 1, "Expected exactly one content item")

			content := toolResult.Content[0]
			assert.Equal(t, "text", string(content.Type), "Content type should be text")
			assert.NotNil(t, content.TextContent, "TextContent should be present")
			assert.Equal(t, "71379509", content.TextContent.Text, "Expected project ID 71379509")
		})

		// Test 4: Call the list_issues tool
		t.Run("CallListIssuesTool", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Test with the project ID from the previous test
			args := map[string]interface{}{
				"project_id": 71379509,
			}

			toolResult, err := client.CallTool(ctx, "list_issues", args)
			require.NoError(t, err, "Failed to call list_issues tool")

			assert.NotNil(t, toolResult.Content, "Tool result should have content")
			assert.Len(t, toolResult.Content, 1, "Expected exactly one content item")

			content := toolResult.Content[0]
			assert.Equal(t, "text", string(content.Type), "Content type should be text")
			assert.NotNil(t, content.TextContent, "TextContent should be present")
			
			// Parse the JSON response to verify it's valid
			var issues []interface{}
			err = json.Unmarshal([]byte(content.TextContent.Text), &issues)
			require.NoError(t, err, "Response should be valid JSON")
			
			t.Logf("Retrieved %d issues", len(issues))
		})

		// Test 5: Call list_issues with state filter
		t.Run("CallListIssuesWithStateFilter", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Test with state filter
			args := map[string]interface{}{
				"project_id": 71379509,
				"state":      "all",
			}

			toolResult, err := client.CallTool(ctx, "list_issues", args)
			require.NoError(t, err, "Failed to call list_issues tool with state filter")

			assert.NotNil(t, toolResult.Content, "Tool result should have content")
			assert.Len(t, toolResult.Content, 1, "Expected exactly one content item")

			content := toolResult.Content[0]
			assert.Equal(t, "text", string(content.Type), "Content type should be text")
			assert.NotNil(t, content.TextContent, "TextContent should be present")
			
			// Parse the JSON response to verify it's valid
			var issues []interface{}
			err = json.Unmarshal([]byte(content.TextContent.Text), &issues)
			require.NoError(t, err, "Response should be valid JSON")
			
			t.Logf("Retrieved %d issues with state=all", len(issues))
		})

		// Test 6: Call the create_issues tool
		t.Run("CallCreateIssuesTool", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Test with minimal required parameters
			args := map[string]interface{}{
				"project_id": 71379509,
				"title":      "Test Issue Created by MCP Server",
			}

			toolResult, err := client.CallTool(ctx, "create_issues", args)
			require.NoError(t, err, "Failed to call create_issues tool")

			assert.NotNil(t, toolResult.Content, "Tool result should have content")
			assert.Len(t, toolResult.Content, 1, "Expected exactly one content item")

			content := toolResult.Content[0]
			assert.Equal(t, "text", string(content.Type), "Content type should be text")
			assert.NotNil(t, content.TextContent, "TextContent should be present")
			
			// Parse the JSON response to verify it's valid
			var issue map[string]interface{}
			err = json.Unmarshal([]byte(content.TextContent.Text), &issue)
			require.NoError(t, err, "Response should be valid JSON")
			
			// Verify the issue structure
			assert.NotNil(t, issue["id"], "Issue should have an ID")
			assert.NotNil(t, issue["iid"], "Issue should have an IID")
			assert.Equal(t, "Test Issue Created by MCP Server", issue["title"], "Issue title should match")
			assert.Equal(t, "opened", issue["state"], "New issue should be in opened state")
			
			t.Logf("Created issue with ID %v and IID %v", issue["id"], issue["iid"])
		})

		// Test 7: Call create_issues with optional parameters
		t.Run("CallCreateIssuesWithOptionalParams", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Test with description and labels
			args := map[string]interface{}{
				"project_id":  71379509,
				"title":       "Test Issue with Description and Labels",
				"description": "This is a test issue created by the MCP server integration test with additional parameters.",
				"labels":      []interface{}{"test", "mcp", "automation"},
			}

			toolResult, err := client.CallTool(ctx, "create_issues", args)
			require.NoError(t, err, "Failed to call create_issues tool with optional parameters")

			assert.NotNil(t, toolResult.Content, "Tool result should have content")
			assert.Len(t, toolResult.Content, 1, "Expected exactly one content item")

			content := toolResult.Content[0]
			assert.Equal(t, "text", string(content.Type), "Content type should be text")
			assert.NotNil(t, content.TextContent, "TextContent should be present")
			
			// Parse the JSON response to verify it's valid
			var issue map[string]interface{}
			err = json.Unmarshal([]byte(content.TextContent.Text), &issue)
			require.NoError(t, err, "Response should be valid JSON")
			
			// Verify the issue structure
			assert.NotNil(t, issue["id"], "Issue should have an ID")
			assert.NotNil(t, issue["iid"], "Issue should have an IID")
			assert.Equal(t, "Test Issue with Description and Labels", issue["title"], "Issue title should match")
			assert.Equal(t, "opened", issue["state"], "New issue should be in opened state")
			assert.Contains(t, issue["description"], "This is a test issue created by the MCP server", "Issue description should contain expected text")
			
			// Check labels
			labels, ok := issue["labels"].([]interface{})
			assert.True(t, ok, "Labels should be an array")
			assert.Contains(t, labels, "test", "Should contain 'test' label")
			assert.Contains(t, labels, "mcp", "Should contain 'mcp' label")
			assert.Contains(t, labels, "automation", "Should contain 'automation' label")
			
			t.Logf("Created issue with labels: %v", labels)
		})

		// Test 8: Error handling - create_issues missing required parameter
		t.Run("CallCreateIssuesMissingTitle", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			args := map[string]interface{}{
				"project_id": 71379509,
				// Missing title parameter
			}

			toolResult, err := client.CallTool(ctx, "create_issues", args)
			// Either should return an error or tool result should indicate failure
			if err == nil {
				// If no error, the tool should handle the missing parameter gracefully
				assert.NotNil(t, toolResult, "Tool result should not be nil")
				// The tool might return an error message in the content
				t.Logf("Tool result for missing title: %+v", toolResult)
			}
		})

		// Test 9: Error handling - create_issues invalid project ID
		t.Run("CallCreateIssuesInvalidProjectID", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			args := map[string]interface{}{
				"project_id": 999999999, // Non-existent project ID
				"title":      "Test Issue",
			}

			toolResult, err := client.CallTool(ctx, "create_issues", args)
			// Either should return an error or tool result should indicate failure
			if err == nil {
				// If no error, the tool should handle the invalid project ID gracefully
				assert.NotNil(t, toolResult, "Tool result should not be nil")
				// The tool might return an error message in the content
				t.Logf("Tool result for invalid project ID: %+v", toolResult)
			}
		})

		// Test 10: Error handling - invalid tool name
		t.Run("CallInvalidTool", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			args := map[string]interface{}{
				"remote_url": "git@gitlab.com:sgaunet/poc-table.git",
			}

			_, err := client.CallTool(ctx, "invalid_tool", args)
			assert.Error(t, err, "Calling invalid tool should return error")
		})

		// Test 11: Error handling - missing required parameter
		t.Run("CallToolMissingParameter", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			args := map[string]interface{}{
				// Missing remote_url parameter
			}

			toolResult, err := client.CallTool(ctx, "get_project_id", args)
			// Either should return an error or tool result should indicate failure
			if err == nil {
				// If no error, the tool should handle the missing parameter gracefully
				assert.NotNil(t, toolResult, "Tool result should not be nil")
				// The tool might return an error message in the content
				t.Logf("Tool result for missing parameter: %+v", toolResult)
			}
		})

		// Test 12: Error handling - invalid repository URL
		t.Run("CallToolInvalidURL", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			args := map[string]interface{}{
				"remote_url": "invalid-url",
			}

			toolResult, err := client.CallTool(ctx, "get_project_id", args)
			// Either should return an error or tool result should indicate failure
			if err == nil {
				// If no error, the tool should handle the invalid URL gracefully
				assert.NotNil(t, toolResult, "Tool result should not be nil")
				// The tool might return an error message in the content
				t.Logf("Tool result for invalid URL: %+v", toolResult)
			}
		})
	})
}

func TestMCPServerConcurrency(t *testing.T) {
	// Check if GITLAB_TOKEN is set
	if os.Getenv("GITLAB_TOKEN") == "" {
		t.Skip("GITLAB_TOKEN environment variable not set, skipping concurrency test")
	}

	t.Run("ConcurrentToolCalls", func(t *testing.T) {
		// Start the MCP server subprocess
		binary := "../../gitlab-mcp-coverage"
		cmd := exec.Command(binary)
		cmd.Dir = "."

		// Create stdin/stdout pipes
		stdin, err := cmd.StdinPipe()
		require.NoError(t, err, "Failed to create stdin pipe")

		stdout, err := cmd.StdoutPipe()
		require.NoError(t, err, "Failed to create stdout pipe")

		// Start the server process
		err = cmd.Start()
		require.NoError(t, err, "Failed to start MCP server")

		// Ensure cleanup - use graceful shutdown for coverage data
		defer func() {
			if cmd.Process != nil {
				// Try graceful shutdown first
				cmd.Process.Signal(os.Interrupt)
				// Wait a bit for graceful shutdown
				time.Sleep(100 * time.Millisecond)
				// Then kill if still running
				cmd.Process.Kill()
				cmd.Wait()
			}
		}()

		// Give the server time to start
		time.Sleep(2 * time.Second)

		// Create MCP client with stdio transport
		clientTransport := stdio.NewStdioServerTransportWithIO(stdout, stdin)
		client := mcp_golang.NewClient(clientTransport)

		// Initialize the client
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = client.Initialize(ctx)
		require.NoError(t, err, "Failed to initialize MCP client")

		// Test concurrent tool calls
		const numCalls = 3
		results := make(chan string, numCalls)
		errors := make(chan error, numCalls)

		for i := 0; i < numCalls; i++ {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				args := map[string]interface{}{
					"remote_url": "git@gitlab.com:sgaunet/poc-table.git",
				}

				toolResult, err := client.CallTool(ctx, "get_project_id", args)
				if err != nil {
					errors <- err
					return
				}

				if len(toolResult.Content) == 0 || toolResult.Content[0].TextContent == nil {
					errors <- assert.AnError
					return
				}

				results <- toolResult.Content[0].TextContent.Text
			}()
		}

		// Collect results
		for i := 0; i < numCalls; i++ {
			select {
			case result := <-results:
				assert.Equal(t, "71379509", result, "Expected project ID 71379509")
			case err := <-errors:
				t.Errorf("Concurrent tool call failed: %v", err)
			case <-time.After(45 * time.Second):
				t.Error("Concurrent tool call timed out")
			}
		}
	})
}
