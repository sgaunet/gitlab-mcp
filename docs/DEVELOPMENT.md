# Development Guide

[< Back to README](../README.md)

Guide for developers contributing to gitlab-mcp.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Architecture](#project-architecture)
- [Build Commands](#build-commands)
- [Running Tests](#running-tests)
- [Code Conventions](#code-conventions)
- [Adding New Tools](#adding-new-tools)
- [CI/CD Pipeline](#cicd-pipeline)
- [Release Process](#release-process)

## Development Setup

### Prerequisites

- **Go 1.21 or later** (check with `go version`)
- **[Task](https://taskfile.dev/)** for build automation
- **Git** for version control
- **Docker** (optional, for container testing)

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/sgaunet/gitlab-mcp.git
cd gitlab-mcp

# Install dependencies
go mod download

# Build the binary
task build

# Or build manually
go build -o gitlab-mcp

# Verify build
./gitlab-mcp --version
```

### Environment Setup

Create a `.env` file for development:

```bash
# .env
GITLAB_TOKEN=your_development_token
GITLAB_URI=https://gitlab.com/
GITLAB_VALIDATE_LABELS=true
```

**Important:** Add `.env` to `.gitignore` to avoid committing secrets.

### Local Testing

```bash
# Run the server locally
export GITLAB_TOKEN=your_token
go run .

# Test with echo pipe
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | go run .
```

## Project Architecture

The codebase follows a clean architecture pattern with clear separation of concerns.

### Directory Structure

```
gitlab-mcp/
├── main.go                  # MCP server entry point, tool registration
├── internal/
│   ├── app/
│   │   ├── app.go           # Business logic for GitLab API integration
│   │   ├── interfaces.go    # Interface definitions for abstraction
│   │   ├── client.go        # Production GitLab client wrapper
│   │   ├── mocks.go         # Mock implementations for testing
│   │   └── app_test.go      # Unit tests
│   └── logger/
│       ├── logger.go        # Structured logging utilities
│       └── logger_test.go   # Logger tests
├── docs/                    # Documentation
├── Taskfile.yml            # Build automation tasks
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
└── Dockerfile              # Container image definition
```

### Key Components

#### main.go

MCP server initialization, protocol setup, and tool registration:
- **Tool Setup Functions**: Each tool has a dedicated `setup*Tool()` function
- **Request Handlers**: Implemented as `handle*Request()` functions with validation
- **Parameter Extraction**: Separated into `extract*Options()` helper functions
- **Pattern**: validate → extract → call app → marshal → return

Example tool registration:
```go
func setupListIssuesTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
    tool := mcp.NewTool("list_issues",
        mcp.WithDescription("..."),
        mcp.WithString("project_path", mcp.Required(), ...),
    )
    s.AddTool(tool, handleListIssuesRequest(appInstance, debugLogger))
}
```

#### internal/app/app.go

Business logic for GitLab API integration:
- **App Struct**: Holds GitLab client, configuration, and logger
- **Public Methods**: Each tool operation (e.g., `ListProjectIssues`, `CreateProjectIssue`)
- **Private Helpers**: Validation and utility functions
- **Interface-Based**: Testability via dependency injection

Example method:
```go
func (a *App) ListProjectIssues(projectPath string, opts *ListIssuesOptions) ([]*gitlab.Issue, error) {
    // Validation
    // API calls
    // Error handling
    // Return results
}
```

#### internal/app/interfaces.go

Interface definitions for GitLab client abstraction:
- **GitLabClient**: Main interface wrapping all service interfaces
- **Service Interfaces**: `ProjectsService`, `IssuesService`, `LabelsService`, etc.
- **Purpose**: Enables mocking in tests without external dependencies

#### internal/app/client.go

Production wrapper implementing GitLabClient interface:
- Wraps the official `gitlab.Client` to match interface contracts
- Provides access to all GitLab service implementations

#### internal/logger/

Structured logging utilities with configurable levels:
- Uses Go's `log/slog` for structured logging
- Outputs to stderr to avoid interfering with MCP stdin/stdout communication

### Key Architectural Patterns

**Dependency Injection**
- `NewWithClient()` allows injecting mocked clients for testing
- Production code uses `New()` which creates a real GitLab client

**Interface Segregation**
- Each GitLab service has its own interface
- Tests only mock the services they need

**Clean Error Handling**
- Static error variables with wrapped context
- Consistent error messages

**MCP Protocol**
- Uses `github.com/mark3labs/mcp-go` library
- JSON-RPC 2.0 over stdio communication

## Build Commands

### Using Task (Recommended)

```bash
# List all available tasks
task --list

# Build the binary
task build

# Build with coverage instrumentation
task build:coverage

# Run linter
task lint

# Run all unit tests
task test

# Show test coverage percentage
task coverage

# Create development snapshot
task snapshot

# Create production release (requires git tag)
task release
```

### Manual Commands

```bash
# Build
go build -o gitlab-mcp

# Build with version info
go build -ldflags "-X main.version=0.9.1" -o gitlab-mcp

# Run tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run linter
golangci-lint run

# Generate coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Running Tests

### Unit Tests

Located in `internal/app/app_test.go` and `internal/logger/logger_test.go`:

```bash
# Run all tests
task test

# Run tests in specific package
go test ./internal/app -v

# Run specific test function
go test ./internal/app -v -run TestValidateConnection

# Run tests with coverage
go test ./internal/app -v -cover
```

### Test Structure

The codebase uses dependency injection via interfaces:

```go
// Create mock client
mockClient := &MockGitLabClient{}
mockProjects := &MockProjectsService{}
mockClient.On("Projects").Return(mockProjects)

// Inject into app
app := NewWithClient(mockClient, cfg, logger)

// Set expectations
mockProjects.On("GetProject", ...).Return(project, nil, nil)

// Test
result, err := app.ListProjectIssues(...)
assert.NoError(t, err)

// Verify
mockProjects.AssertExpectations(t)
```

### Coverage Analysis

```bash
# Show coverage percentage
task coverage

# Generate HTML report
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# View coverage by function
go tool cover -func=coverage.out
```

Current test coverage: **72.7%**

### Manual Testing

See [MANUAL_TEST.md](../MANUAL_TEST.md) for step-by-step manual testing instructions.

## Code Conventions

### Go Standards

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` for formatting (enforced by CI)
- Run `golangci-lint` before committing
- Add comments to exported functions and types

### Error Handling

Use static error variables:
```go
var (
    ErrProjectPathRequired = errors.New("project_path is required")
    ErrInvalidStateValue   = errors.New("state must be 'opened' or 'closed'")
)
```

Wrap errors with context:
```go
if err != nil {
    return nil, fmt.Errorf("failed to get project: %w", err)
}
```

### Logging

Use structured logging with `slog`:
```go
debugLogger.Debug("Processing request",
    "project_path", projectPath,
    "state", opts.State,
)

debugLogger.Error("API call failed",
    "error", err,
    "project_path", projectPath,
)
```

### Testing

- Write tests for all public methods
- Use table-driven tests for multiple scenarios
- Mock external dependencies (GitLab API)
- Test error paths, not just happy paths

Example table-driven test:
```go
tests := []struct {
    name    string
    input   string
    wantErr bool
}{
    {"valid path", "myorg/myproject", false},
    {"empty path", "", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test logic
    })
}
```

## Adding New Tools

Follow this pattern when adding a new MCP tool:

### 1. Define the tool in main.go

```go
func setupNewTool(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
    newTool := mcp.NewTool("tool_name",
        mcp.WithDescription("Tool description"),
        mcp.WithString("param_name", mcp.Required(), mcp.Description("param description")),
    )
    s.AddTool(newTool, handleNewToolRequest(appInstance, debugLogger))
}
```

### 2. Implement the handler function

```go
func handleNewToolRequest(appInstance *app.App, debugLogger *slog.Logger) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        args := request.GetArguments()
        debugLogger.Debug("Received tool request", "args", args)

        // Extract and validate parameters
        paramValue, ok := args["param_name"].(string)
        if !ok || paramValue == "" {
            return mcp.NewToolResultError("param_name is required"), nil
        }

        // Call app method
        result, err := appInstance.NewMethod(paramValue)
        if err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }

        // Marshal response to JSON
        resultJSON, err := json.Marshal(result)
        if err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
        }

        return mcp.NewToolResultText(string(resultJSON)), nil
    }
}
```

### 3. Add business logic to internal/app/app.go

```go
// Add method to App struct
func (a *App) NewMethod(paramValue string) (*ResultType, error) {
    // Validate inputs
    if paramValue == "" {
        return nil, ErrParamRequired
    }

    // Make API calls using a.client
    result, _, err := a.client.SomeService().SomeMethod(paramValue)
    if err != nil {
        return nil, fmt.Errorf("failed to call API: %w", err)
    }

    return result, nil
}
```

### 4. Add service interface (if needed)

If using a new GitLab service, add to `internal/app/interfaces.go`:

```go
type NewService interface {
    SomeMethod(param string, opt *gitlab.Options) (*gitlab.Result, *gitlab.Response, error)
}

type GitLabClient interface {
    // ... existing methods
    NewService() NewService
}
```

And implement in `internal/app/client.go`:

```go
func (c *Client) NewService() NewService {
    return c.client.NewService
}
```

### 5. Write tests

Add tests in `internal/app/app_test.go`:

```go
func TestNewMethod(t *testing.T) {
    // Setup
    mockClient := &MockGitLabClient{}
    mockService := &MockNewService{}
    mockClient.On("NewService").Return(mockService)

    app := NewWithClient(mockClient, &Config{}, logger)

    // Expectations
    expected := &ResultType{/* ... */}
    mockService.On("SomeMethod", "test").Return(expected, nil, nil)

    // Execute
    result, err := app.NewMethod("test")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    mockService.AssertExpectations(t)
}
```

### 6. Register in main()

Add call to `setup*Tool()` in `registerAllTools()`:

```go
func registerAllTools(s *server.MCPServer, appInstance *app.App, debugLogger *slog.Logger) {
    // ... existing tools
    setupNewTool(s, appInstance, debugLogger)
}
```

### 7. Update documentation

- Add tool to README.md tool list
- Create detailed documentation in docs/TOOLS.md
- Add usage examples
- Update CLAUDE.md if needed

## CI/CD Pipeline

### GitHub Actions Workflows

Located in `.github/workflows/`:

**Coverage (`coverage.yml`)**
- Runs on every push and PR
- Executes unit tests with coverage tracking
- Updates coverage badge in wiki

**Snapshot Build (`snapshot.yml`)**
- Triggers on push to main branch
- Creates development snapshots
- Publishes to GitHub Container Registry

**Release Build (`release.yml`)**
- Triggers on version tags (e.g., `v0.9.1`)
- Builds multi-architecture binaries
- Creates GitHub release
- Publishes Docker images
- Updates Homebrew tap

**Vulnerability Scan (`vulnerability-scan.yml`)**
- Runs security scanning with Trivy
- Checks dependencies for known vulnerabilities
- Runs on push and scheduled

### Running CI Checks Locally

```bash
# Run linter (same as CI)
task lint

# Run tests with coverage (same as CI)
task test
task coverage

# Build snapshot (similar to CI)
task snapshot
```

## Release Process

### Creating a Release

1. **Update version information:**
   ```bash
   # Ensure all changes are committed
   git status

   # Create and push version tag
   git tag v0.10.0
   git push origin v0.10.0
   ```

2. **Wait for CI/CD:**
   - GitHub Actions will automatically:
     - Build multi-architecture binaries
     - Create Docker images
     - Publish to GitHub Container Registry
     - Create GitHub release with binaries
     - Update Homebrew tap

3. **Verify release:**
   - Check GitHub releases page
   - Test Docker image: `docker pull ghcr.io/sgaunet/gitlab-mcp:0.10.0`
   - Test Homebrew: `brew upgrade sgaunet/tools/gitlab-mcp`

### Versioning

Follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking API changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Release Checklist

- [ ] All tests passing
- [ ] Code reviewed and merged to main
- [ ] Documentation updated
- [ ] CHANGELOG updated (if applicable)
- [ ] Version tag created
- [ ] CI/CD pipeline succeeds
- [ ] Release notes added on GitHub
- [ ] Docker image tested
- [ ] Homebrew formula tested

## Contributing Guidelines

See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Code review process
- Commit message conventions
- Issue reporting guidelines
- Pull request workflow

## Getting Help

- **Issues**: [github.com/sgaunet/gitlab-mcp/issues](https://github.com/sgaunet/gitlab-mcp/issues)
- **Discussions**: GitHub Discussions (if available)
- **Documentation**: [docs/](.)

## Next Steps

- [Setup guide →](SETUP.md)
- [Tool reference →](TOOLS.md)
- [Contributing guidelines →](CONTRIBUTING.md)
