# GitLab MCP Server

A Model Context Protocol (MCP) server that provides GitLab integration tools for Claude Code.

## Features

- **List Issues**: List issues for a GitLab project using project path (namespace/project-name)
- **Create Issues**: Create new issues with title, description, labels, and assignees
- **List Labels**: List project labels with optional filtering and counts
- Direct project path access - no need to resolve project IDs
- Compatible with Claude Code's MCP architecture

## Prerequisites

- Go 1.21 or later
- [Task](https://taskfile.dev/) for build automation
- Claude Code CLI
- Access to GitLab repositories

## Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd gitlab-mcp
   ```

2. **Build the project:**
   ```bash
   task build
   ```
   
   Or manually:
   ```bash
   go build -o gitlab-mcp
   ```

3. **Add to Claude Code:**
   ```bash
   claude mcp add gitlab-mcp -s user -- /path/to/gitlab-mcp
   ```
   
   Replace `/path/to/gitlab-mcp` with the actual path to your built executable.

## Usage

Once installed, the MCP server provides the following tools in Claude Code:

### list_issues

Lists issues for a GitLab project using the project path.

**Parameters:**
- `project_path` (string, required): GitLab project path (e.g., 'namespace/project-name')
- `state` (string, optional): Filter by issue state (`opened`, `closed`, `all`) - defaults to `opened`
- `labels` (string, optional): Comma-separated list of labels to filter by
- `limit` (number, optional): Maximum number of issues to return (default: 100, max: 100)

**Examples:**
```
List all open issues for project sgaunet/poc-table
```

```
List all issues (open and closed) for project sgaunet/poc-table
```

```
List issues with state=all and limit=50 for project sgaunet/poc-table
```

**Response Format:**
Returns a JSON array of issue objects, each containing:
- `id`: Issue ID
- `iid`: Internal issue ID
- `title`: Issue title
- `description`: Issue description
- `state`: Issue state (`opened` or `closed`)
- `labels`: Array of label names
- `assignees`: Array of assignee objects
- `created_at`: Creation timestamp
- `updated_at`: Last update timestamp

### create_issues

Creates a new issue for a GitLab project.

**Parameters:**
- `project_path` (string, required): GitLab project path (e.g., 'namespace/project-name')
- `title` (string, required): Issue title
- `description` (string, optional): Issue description
- `labels` (array, optional): Array of labels to assign to the issue
- `assignees` (array, optional): Array of user IDs to assign to the issue

**Examples:**
```
Create an issue with title "Bug fix needed" for project sgaunet/poc-table
```

```
Create an issue with title "Feature request", description "Add new functionality", and labels ["enhancement", "feature"] for project sgaunet/poc-table
```

**Response Format:**
Returns a JSON object of the created issue with the same structure as list_issues.

### list_labels

Lists labels for a GitLab project with optional filtering.

**Parameters:**
- `project_path` (string, required): GitLab project path (e.g., 'namespace/project-name')
- `with_counts` (boolean, optional): Include issue and merge request counts (default: false)
- `include_ancestor_groups` (boolean, optional): Include labels from ancestor groups (default: false)
- `search` (string, optional): Filter labels by search keyword
- `limit` (number, optional): Maximum number of labels to return (default: 100, max: 100)

**Examples:**
```
List all labels for project sgaunet/poc-table
```

```
List labels with counts for project sgaunet/poc-table
```

```
Search for labels containing "bug" in project sgaunet/poc-table
```

**Response Format:**
Returns a JSON array of label objects, each containing:
- `id`: Label ID
- `name`: Label name
- `color`: Label color (hex code)
- `text_color`: Text color for the label
- `description`: Label description
- `open_issues_count`: Number of open issues (if with_counts=true)
- `closed_issues_count`: Number of closed issues (if with_counts=true)
- `open_merge_requests_count`: Number of open merge requests (if with_counts=true)

## Development

### Build Tasks

This project uses [Task](https://taskfile.dev/) for build automation. Available tasks:

```bash
# List all available tasks
task --list

# Build the binary
task build

# Build with coverage support
task build:coverage

# Run linter
task linter

# Run integration tests
task test:integration

# Run integration tests with coverage
task test:integration:coverage

# Calculate coverage from collected data
task coverage:calculate

# Clean coverage data
task coverage:clean

# Complete coverage workflow
task coverage:full
```

### Running Tests

**Unit Tests:**
```bash
go test ./...
```

**Integration Tests:**
```bash
task test:integration
```

**Coverage Testing:**
```bash
task coverage:full
```

The coverage workflow builds a coverage-enabled binary, runs integration tests, and provides detailed coverage statistics:
- Main package coverage
- Internal/app package coverage  
- Internal/logger package coverage

### Manual Testing

See [MANUAL_TEST.md](MANUAL_TEST.md) for step-by-step manual testing instructions.

### Automated Testing

Run the test script:
```bash
./test_working.sh
```

## Configuration

The MCP server runs as a subprocess and communicates via JSON-RPC over stdin/stdout. No additional configuration is required.

## Troubleshooting

### Common Issues

1. **Parse Error (-32700)**: Ensure you're using the server through Claude Code's MCP interface, not directly inputting raw URLs
2. **Invalid URL**: Verify the GitLab URL format is correct (SSH or HTTPS)
3. **Connection Issues**: Check network connectivity to GitLab

### Debug Mode

The server outputs debug information to stderr, which can be helpful for troubleshooting:
```bash
go run . < input.txt 2> debug.log
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License. See [LICENSE](LICENSE) for details.

## Support

For issues and questions, please create an issue in the repository.