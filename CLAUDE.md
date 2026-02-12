# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Model Context Protocol (MCP) server that provides GitLab integration tools for Claude Code. It's implemented as a Go application that communicates via JSON-RPC 2.0 over stdin/stdout, providing tools for GitLab project management:

- `list_issues`: Lists issues for a GitLab project using project path with filtering options
- `create_issues`: Creates new issues with title, description, labels (validated), and assignees
- `update_issues`: Updates existing issues (title, description, state, labels (validated), assignees)
- `list_labels`: Lists project labels with optional filtering and counts
- `add_issue_note`: Adds notes/comments to existing issues
- `get_project_description`: Gets the current description of a GitLab project
- `update_project_description`: Updates the description of a GitLab project
- `get_project_topics`: Gets the current topics/tags of a GitLab project
- `update_project_topics`: Updates the topics/tags of a GitLab project (replaces all existing topics)
- `list_epics`: Lists epics for a GitLab group (Premium/Ultimate tier)
- `create_epic`: Creates new epics in a GitLab group with title, description, labels (validated), dates, and confidentiality (Premium/Ultimate tier)
- `get_latest_pipeline`: Gets the latest pipeline for a GitLab project with optional ref (branch/tag) filtering
- `list_pipeline_jobs`: Lists all jobs for a GitLab pipeline with filtering options (status, stage)
- `get_job_log`: Retrieves complete log output for a specific CI/CD job with job metadata
- `download_job_trace`: Downloads CI/CD job logs to local files for offline analysis and archiving

## Architecture

The codebase follows a clean architecture pattern with clear separation of concerns:

### Main Components
- **main.go**: MCP server initialization, protocol setup, and tool registration
  - Each tool has a dedicated `setup*Tool()` function that creates the tool definition
  - Request handlers are implemented as `handle*Request()` functions with input validation
  - Parameter extraction is separated into `extract*Options()` helper functions
  - All tools follow the same pattern: validate → extract → call app → marshal → return

- **internal/app/app.go**: Business logic for GitLab API integration
  - `App` struct holds GitLab client, configuration, and logger
  - Public methods for each tool operation (e.g., `ListProjectIssues`, `CreateProjectIssue`)
  - Private helper methods for validation
  - Interface-based design for testability via dependency injection

- **internal/app/interfaces.go**: Interface definitions for GitLab client abstraction
  - Defines service interfaces matching GitLab API structure
  - Enables mocking in tests without external dependencies

- **internal/app/client.go**: Production wrapper implementing GitLabClient interface
  - Wraps the official `gitlab.Client` to match interface contracts
  - Provides access to all GitLab service implementations

- **internal/logger**: Structured logging utilities with configurable levels
  - Uses Go's `log/slog` for structured logging
  - Outputs to stderr to avoid interfering with MCP stdin/stdout communication

### Key Architectural Patterns
- **Dependency Injection**: `NewWithClient()` allows injecting mocked clients for testing
- **Interface Segregation**: Each GitLab service (Issues, Projects, Labels, etc.) has its own interface
- **Clean Error Handling**: Static error variables with wrapped context in error messages
- **MCP Protocol**: Uses `github.com/mark3labs/mcp-go` library for JSON-RPC 2.0 over stdio

The server uses the official GitLab Go client library (`gitlab.com/gitlab-org/api/client-go`) and supports both GitLab.com and self-hosted instances.

### Usage Examples

**Listing issues:**

By default, the `list_issues` tool returns issues from both the project and its parent group(s) for comprehensive visibility:

```
List issues for myorg/team/project
```

This returns issues from:
- The project `myorg/team/project`
- The parent group `myorg/team` (and all its projects)

**Excluding group issues:**

To retrieve only project-level issues:

```
List issues for myorg/project with include_group_issues=false
```

**How deduplication works:**

When group issues are included, issues that belong to the current project are deduplicated based on their ProjectID. This ensures each issue appears only once in the results.

**Graceful fallback:**

If fetching group issues fails (due to permissions or other errors), the tool automatically falls back to returning only project issues, ensuring the request always succeeds when project access is available.

**Managing project descriptions:**
```
Get the description of sgaunet/gitlab-mcp
```

```
Update the description of myorg/myproject to "A comprehensive GitLab integration tool for MCP"
```

**Managing project topics:**
```
Get the topics for myorg/myproject
```

```
Update the topics of myorg/myproject to ["golang", "gitlab", "mcp", "api", "automation"]
```

```
Remove all topics from myorg/myproject by setting topics to []
```

**Managing epics (Premium/Ultimate tier):**
```
List epics for the myorg group
```

```
Create an epic in myorg group with title "Q1 2024 Launch"
```

```
Create an epic in myorg/platform group with title "Authentication Redesign", description "Modernize auth with OAuth2 and JWT", labels ["security", "high-priority"], start date "2024-03-01", due date "2024-06-30", and make it confidential
```

**Updating epics:**
```
Update the title of epic 5 in myorg group to "Q2 2024 Launch"
```

```
Update epic 3 in myorg/platform with description "Updated roadmap", state "closed", and labels ["completed", "archived"]
```

```
Update epic 2 dates: start "2024-04-01", due "2024-06-30", and make it confidential
```

```
Update only the labels on epic 7 to ["backend", "infrastructure"]
```

**Debugging CI/CD pipelines:**
```
Get the latest pipeline for myorg/myproject
```

```
List all jobs for the latest pipeline in myorg/myproject
```

```
List only failed jobs for the latest pipeline on the main branch in myorg/myproject
```

```
List failed jobs in the test stage for pipeline ID 12345 in myorg/myproject
```

```
Get the log for job 12345 in pipeline 999 in myorg/myproject
```

```
Get the log for job 54321 from the latest pipeline in myorg/myproject
```

```
Get the log for job 11111 from the latest pipeline on develop branch in myorg/myproject
```

**Downloading job logs:**
```
Download the log for job 12345 to a file
```

```
Download the log for job 54321 from the latest pipeline to ./logs/job_54321.log
```

```
Download the log for job 11111 from the develop branch to /tmp/build.log
```

The tool handles the resolution automatically - no need to look up user IDs or milestone IDs manually.

## CLI Flags for Token Optimization

The server supports CLI flags to disable tool categories, reducing token consumption for specialized use cases:

### Available Flags

- `--no-issues` - Disable issue management tools (4 tools)
- `--no-labels` - Disable label management tools (1 tool)
- `--no-project-metadata` - Disable project metadata tools (4 tools)
- `--no-epics` - Disable epic management tools (3 tools)
- `--no-pipelines` - Disable CI/CD pipeline tools (4 tools)

### Usage Examples

**CI/CD debugging agent (only pipeline tools):**
```bash
gitlab-mcp --no-issues --no-labels --no-project-metadata --no-epics
```

**Documentation agent (only project metadata):**
```bash
gitlab-mcp --no-issues --no-labels --no-epics --no-pipelines
```

**Issue triage bot (only issues and labels):**
```bash
gitlab-mcp --no-project-metadata --no-epics --no-pipelines
```

### Configuration with Claude Code

In `.mcp.json`:
```json
{
  "mcpServers": {
    "gitlab-mcp": {
      "type": "stdio",
      "command": "/usr/local/bin/gitlab-mcp",
      "args": ["--no-epics", "--no-pipelines"],
      "env": {
        "GITLAB_TOKEN": "your_token"
      }
    }
  }
}
```

**Note:** All tools are enabled by default. Use flags to opt-out of categories you don't need.

## Common Development Commands

### Build and Test
```bash
# Build the binary
task build

# Build with coverage support
task build:coverage

# Run linter
task lint

# Run all unit tests
task test

# Run tests in a specific package
go test ./internal/app -v

# Run a specific test function
go test ./internal/app -v -run TestValidateConnection

# Run tests with coverage details
go test ./internal/app -v -cover
```

### Coverage Analysis
```bash
# Show test coverage percentage
task coverage

# Generate HTML coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Release
```bash
# Create development snapshot
task snapshot

# Create production release
task release
```

### Local Development
```bash
# Run the server locally (requires GITLAB_TOKEN)
export GITLAB_TOKEN=your_token_here
export GITLAB_URI=https://gitlab.com/  # optional
go run .
```

## Testing Architecture

### Unit Tests
Located in `internal/app/app_test.go` and `internal/logger/logger_test.go`, these tests:
- Use interface abstractions and mocking for fast, reliable testing
- Test all core functionality without external dependencies
- Include comprehensive coverage of happy paths, error scenarios, and edge cases
- Run in milliseconds with no network or GitLab API requirements

### Test Structure
The codebase uses dependency injection via interfaces for testability:
- **Interfaces**: Defined in `internal/app/interfaces.go` for GitLab client abstraction
  - `GitLabClient`: Main interface wrapping all GitLab service interfaces
  - Service-specific interfaces: `ProjectsService`, `IssuesService`, `LabelsService`, `UsersService`, `NotesService`, `MergeRequestsService`, `MilestonesService`
- **Production Implementation**: `internal/app/client.go` wraps the real GitLab client
- **Mocks**: Generated in `internal/app/mocks.go` using testify/mock
- **Test Constructor**: Use `NewWithClient()` in tests to inject mocked GitLabClient
- **Coverage**: Standard Go coverage tools with HTML report generation

### Adding New Tests
When adding new functionality:
1. Define service interfaces in `internal/app/interfaces.go` if needed
2. Create mock implementations in `internal/app/mocks.go`
3. Write tests using `NewWithClient()` to inject mocks
4. Use `testify/mock` for setting expectations and verifying calls

## Environment Configuration

### Required
- `GITLAB_TOKEN`: GitLab API personal access token with appropriate scopes (`api`, `read_api`, `write_api`)

### Optional
- `GITLAB_URI`: GitLab instance URI (defaults to `https://gitlab.com/`)
- `GITLAB_VALIDATE_LABELS`: Enable/disable label validation for issue and epic creation/updates (defaults to `true`)
  - `true`: Validates that labels exist in the project/group before creating/updating issues/epics (prevents typos)
  - `false`: Allows creating/updating issues/epics with non-existent labels (GitLab API ignores invalid labels)

#### Label Validation Hierarchy

When `GITLAB_VALIDATE_LABELS=true` (default), labels are validated hierarchically to encourage reuse across the project hierarchy:

1. **Project labels** are checked first
2. **Parent group labels** are checked
3. **Grandparent group labels** (and so on up the hierarchy)

This hierarchical validation encourages label reuse and reduces label proliferation by making group-level labels available to all child projects.

**Example**:
- Root group `myorg` has labels: `["roadmap", "high-priority"]`
- Sub-group `myorg/team` has labels: `["backend", "frontend"]`
- Project `myorg/team/project` has labels: `["bug", "feature"]`
- Creating an issue with `["bug", "roadmap", "backend"]` succeeds because:
  - `"bug"` exists at project level
  - `"backend"` exists at parent group level
  - `"roadmap"` exists at root group level

**Benefits**:
- Encourages consistent labeling across projects within an organization
- Reduces duplicate labels (e.g., multiple projects creating their own "high-priority" label)
- Follows GitLab's best practice of defining labels at the appropriate group level
- Provides clear error messages showing all available labels from the entire hierarchy

**Note**: The `list_labels` tool also supports hierarchical label listing by default. Set `include_ancestor_groups=false` to view only project-specific labels.

### Go Version
- Requires Go 1.24.3 or later (as specified in `go.mod`)

## Key Implementation Details

### Project Path Processing
The app uses GitLab project paths in the format `namespace/project-name` (e.g., `namespace/project_name`) to directly access projects via the GitLab API's `GetProject()` method. This eliminates the need for URL parsing and project ID resolution.

### GitLab API Integration
The server uses the official GitLab Go client to interact with GitLab's REST API:
- **Project Access**: `/projects/:path` endpoint for direct project access by path
- **Issues List**: `/projects/:id/issues` endpoint for project issue retrieval
- **Group Issues List**: `/groups/:id/issues` endpoint for group issue retrieval (merged with project issues by default)
- **Issue Creation**: `/projects/:id/issues` endpoint for creating new issues
- **Issue Updates**: `/projects/:id/issues/:issue_iid` endpoint for updating existing issues
- **Notes Creation**: `/projects/:id/issues/:issue_iid/notes` endpoint for adding notes/comments to issues
- **Labels List**: `/projects/:id/labels` endpoint for label retrieval
- **Filtering**: Supports state filtering, label filtering, search, and pagination
- **Deduplication**: Group issues are filtered to exclude duplicates from the current project
- **Response Format**: Returns structured JSON with issue, label, note, and project metadata

### MCP Protocol Implementation
- Uses `github.com/mark3labs/mcp-go` for server implementation
- Implements MCP protocol version "2025-03-26"
- Advertises tools and resources capabilities
- Handles JSON-RPC 2.0 communication over stdio

### Error Handling
- Comprehensive error handling with context preservation
- Graceful degradation for missing/invalid configurations
- Structured logging for debugging and monitoring

## Testing Coverage

Unit tests provide comprehensive coverage of all functionality:
- `ValidateConnection`: Token validation and error handling
- `ListProjectIssues`: Issue retrieval with state filtering, parameter validation, and error scenarios
- `CreateProjectIssue`: Issue creation with various parameter combinations and validation
- `UpdateProjectIssue`: Issue updates with partial/full updates, state changes, and validation
- `AddIssueNote`: Note creation with body validation, project resolution, and API error handling
- `ListProjectLabels`: Label retrieval with optional filtering, search, and counts
- `GetLatestPipeline`: Pipeline retrieval with ref filtering and error scenarios
- `ListPipelineJobs`: Job listing with pipeline ID resolution, status/stage filtering, and comprehensive error handling
- Edge cases: nil parameters, empty values, API errors, project not found, invalid issue IIDs

All tests run without external dependencies using mocked GitLab client interfaces. Current test coverage is **72.7%**.

### Group Issues Integration

The `list_issues` tool includes enhanced functionality for retrieving issues from parent groups:

**Architecture:**
- Uses `ListGroupIssues` API endpoint (`/groups/:id/issues`) to fetch group-level issues
- Automatically extracts group path from project path (e.g., `myorg/team/project` → `myorg/team`)
- Merges project and group issues with deduplication based on `ProjectID`
- Graceful fallback to project-only results if group fetching fails

**Default Behavior:**
- `include_group_issues` parameter defaults to `true` for comprehensive results
- Users can opt-out by setting `include_group_issues=false`

**Error Scenarios Handled:**
- Invalid project paths (missing group): Falls back to project-only results
- Group access denied: Falls back to project-only results
- Group not found: Falls back to project-only results
- Permission errors: Falls back to project-only results

**Test Coverage:**
- Helper functions: `extractGroupPath()`, `mergeIssues()`
- Integration tests: merged results, deduplication, graceful fallbacks
- Edge cases: empty sets, full/partial overlap, permission errors

## Adding New Tools

When adding a new MCP tool to the server, follow this pattern:

1. **Define the tool in main.go**:
   ```go
   func setupNewTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
       newTool := mcp.NewTool("tool_name",
           mcp.WithDescription("Tool description"),
           mcp.WithString("param_name", mcp.Required(), mcp.Description("param description")),
       )
       s.AddTool(newTool, handleNewToolRequest(appInstance, debugLogger))
   }
   ```

2. **Implement the handler function**:
   ```go
   func handleNewToolRequest(appInstance *app.App, debugLogger *slog.Logger) func(...) {
       return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
           // Extract and validate parameters
           // Call app method
           // Marshal response to JSON
           // Return result
       }
   }
   ```

3. **Add business logic to internal/app/app.go**:
   - Create public method on `App` struct
   - Add any necessary interfaces to `internal/app/interfaces.go`
   - Implement using the injected `GitLabClient`

4. **Write tests in internal/app/app_test.go**:
   - Mock the required service interfaces
   - Test happy path, error cases, and edge cases
   - Use `NewWithClient()` to inject mocks

5. **Register in main()**:
   - Add call to `setup*Tool()` in `registerAllTools()`

## Task Master AI Instructions
**Import Task Master's development workflow commands and guidelines, treat as if import is in the main CLAUDE.md file.**
@./.taskmaster/CLAUDE.md
