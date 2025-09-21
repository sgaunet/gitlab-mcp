[![GitHub release](https://img.shields.io/github/release/sgaunet/gitlab-mcp.svg)](https://github.com/sgaunet/gitlab-mcp/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/sgaunet/gitlab-mcp)](https://goreportcard.com/report/github.com/sgaunet/gitlab-mcp)
![GitHub Downloads](https://img.shields.io/github/downloads/sgaunet/gitlab-mcp/total)
![Coverage](https://raw.githubusercontent.com/wiki/sgaunet/gitlab-mcp/coverage-badge.svg)
[![coverage](https://github.com/sgaunet/gitlab-mcp/actions/workflows/coverage.yml/badge.svg)](https://github.com/sgaunet/gitlab-mcp/actions/workflows/coverage.yml)
[![Snapshot Build](https://github.com/sgaunet/gitlab-mcp/actions/workflows/snapshot.yml/badge.svg)](https://github.com/sgaunet/gitlab-mcp/actions/workflows/snapshot.yml)
[![Release Build](https://github.com/sgaunet/gitlab-mcp/actions/workflows/release.yml/badge.svg)](https://github.com/sgaunet/gitlab-mcp/actions/workflows/release.yml)
[![Vulnerability Scan](https://github.com/sgaunet/gitlab-mcp/actions/workflows/vulnerability-scan.yml/badge.svg)](https://github.com/sgaunet/gitlab-mcp/actions/workflows/vulnerability-scan.yml)
[![License](https://img.shields.io/github/license/sgaunet/gitlab-mcp.svg)](LICENSE)

# GitLab MCP Server

A Model Context Protocol (MCP) server that provides GitLab integration tools for Claude Code.

## Features

- **List Issues**: List issues for a GitLab project using project path (namespace/project-name)
- **Create Issues**: Create new issues with title, description, labels, and assignees
- **Update Issues**: Update existing issues (title, description, state, labels, assignees)
- **List Labels**: List project labels with optional filtering and counts
- **Add Issue Notes**: Add comments/notes to existing issues
- **Create Merge Requests**: Create new merge requests with source/target branches, title, description, assignees, reviewers, and labels
- **Get Project Description**: Retrieve the current description of a GitLab project
- **Update Project Description**: Update the description of a GitLab project
- **Get Project Topics**: Retrieve the current topics/tags of a GitLab project
- **Update Project Topics**: Update the topics/tags of a GitLab project (replaces all existing topics)
- Direct project path access - no need to resolve project IDs
- Compatible with Claude Code's MCP architecture

## Prerequisites

- Go 1.21 or later
- [Task](https://taskfile.dev/) for build automation
- Claude Code CLI
- Access to GitLab repositories

## Installation

### Option 1: Install with Homebrew (Recommended for macOS/Linux)

```bash
# Add the tap and install
brew tap sgaunet/homebrew-tools
brew install sgaunet/tools/gitlab-mcp
```

### Option 2: Download from GitHub Releases

1. **Download the latest release:**
   
   Visit the [releases page](https://github.com/sgaunet/gitlab-mcp/releases/latest) and download the appropriate binary for your platform:
   
   - **macOS**: `gitlab-mcp_VERSION_darwin_amd64` (Intel) or `gitlab-mcp_VERSION_darwin_arm64` (Apple Silicon)
   - **Linux**: `gitlab-mcp_VERSION_linux_amd64` (x86_64) or `gitlab-mcp_VERSION_linux_arm64` (ARM64)
   - **Windows**: `gitlab-mcp_VERSION_windows_amd64.exe`

2. **Make it executable (macOS/Linux):**
   ```bash
   chmod +x gitlab-mcp_*
   ```

3. **Move to a location in your PATH:**
   ```bash
   # Example for macOS/Linux
   sudo mv gitlab-mcp_* /usr/local/bin/gitlab-mcp
   ```

### Option 3: Build from Source

1. **Clone the repository:**
   ```bash
   git clone https://github.com/sgaunet/gitlab-mcp.git
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

3. **Install to your PATH:**
   ```bash
   sudo mv gitlab-mcp /usr/local/bin/
   ```

### Configuration

The MCP server requires a GitLab personal access token (PAT) with appropriate scopes (e.g., `api`, `read_api`, `write_api`) to interact with the GitLab API.
Set the `GITLAB_TOKEN` environment variable with your PAT:

```bash
export GITLAB_TOKEN=your_personal_access_token
```
You can also set the `GITLAB_URI` environment variable if you're using a self-hosted GitLab instance (default is `https://gitlab.com`):

```bash
export GITLAB_URI=https://your.gitlab.instance
```

The easiest way to set these variables permanently is to add them to your shell profile (e.g., `~/.bashrc`, `~/.zshrc`). It avoids issues with environment variables not being available when Claude Code starts the MCP server.

### Add to Claude Code

After installation, add the MCP server to Claude Code:

```bash
# If installed via Homebrew (Apple Silicon)
claude mcp add gitlab-mcp -s user -- /opt/homebrew/bin/gitlab-mcp

# If installed via Homebrew (Intel Mac) or manually to /usr/local/bin
claude mcp add gitlab-mcp -s user -- /usr/local/bin/gitlab-mcp

# If installed elsewhere, adjust the path accordingly
claude mcp add gitlab-mcp -s user -- /path/to/gitlab-mcp
```

## Usage

Once installed, the MCP server provides the following tools in Claude Code:

### list_issues

Lists issues for a GitLab project using the project path.

**Parameters:**
- `project_path` (string, required): GitLab project path including all namespaces (e.g., 'namespace/project-name' or 'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL
- `state` (string, optional): Filter by issue state (`opened`, `closed`, `all`) - defaults to `opened`
- `labels` (string, optional): Comma-separated list of labels to filter by
- `limit` (number, optional): Maximum number of issues to return (default: 100, max: 100)

**Examples:**
```
List all open issues for project namespace/project_name
```

```
List all issues (open and closed) for project namespace/project_name
```

```
List issues with state=all and limit=50 for project namespace/project_name
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
- `project_path` (string, required): GitLab project path including all namespaces (e.g., 'namespace/project-name' or 'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL
- `title` (string, required): Issue title
- `description` (string, optional): Issue description
- `labels` (array, optional): Array of labels to assign to the issue
- `assignees` (array, optional): Array of user IDs to assign to the issue

**Examples:**
```
Create an issue with title "Bug fix needed" for project namespace/project_name
```

```
Create an issue with title "Feature request", description "Add new functionality", and labels ["enhancement", "feature"] for project namespace/project_name
```

**Response Format:**
Returns a JSON object of the created issue with the same structure as list_issues.

### update_issues

Updates an existing issue for a GitLab project.

**Parameters:**
- `project_path` (string, required): GitLab project path including all namespaces (e.g., 'namespace/project-name' or 'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL
- `issue_iid` (number, required): Issue internal ID (IID) to update
- `title` (string, optional): Updated issue title
- `description` (string, optional): Updated issue description
- `state` (string, optional): Issue state ('opened' or 'closed')
- `labels` (array, optional): Array of labels to assign to the issue
- `assignees` (array, optional): Array of user IDs to assign to the issue

**Examples:**
```
Update the title of issue #5 for project namespace/project_name
```

```
Close issue #10 and update its description for project namespace/project_name
```

**Response Format:**
Returns a JSON object of the updated issue with the same structure as list_issues.

### list_labels

Lists labels for a GitLab project with optional filtering.

**Parameters:**
- `project_path` (string, required): GitLab project path including all namespaces (e.g., 'namespace/project-name' or 'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL
- `with_counts` (boolean, optional): Include issue and merge request counts (default: false)
- `include_ancestor_groups` (boolean, optional): Include labels from ancestor groups (default: false)
- `search` (string, optional): Filter labels by search keyword
- `limit` (number, optional): Maximum number of labels to return (default: 100, max: 100)

**Examples:**
```
List all labels for project namespace/project_name
```

```
List labels with counts for project namespace/project_name
```

```
Search for labels containing "bug" in project namespace/project_name
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

### add_issue_note

Adds a note/comment to an existing issue for a GitLab project.

**Parameters:**
- `project_path` (string, required): GitLab project path including all namespaces (e.g., 'namespace/project-name' or 'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL
- `issue_iid` (number, required): Issue internal ID (IID) to add note to
- `body` (string, required): Note/comment body text

**Examples:**
```
Add a comment "This looks good to me!" to issue #5 for project namespace/project_name
```

```
Add a note "Fixed in latest commit" to issue #12 for project namespace/project_name
```

**Response Format:**
Returns a JSON object of the created note containing:
- `id`: Note ID
- `body`: Note body text
- `author`: Author object with id, username, and name
- `created_at`: Creation timestamp
- `updated_at`: Last update timestamp
- `system`: Boolean indicating if this is a system-generated note
- `noteable`: Object containing information about the issue this note belongs to

### create_merge_request

Creates a new merge request for a GitLab project.

**Parameters:**
- `project_path` (string, required): GitLab project path including all namespaces (e.g., 'namespace/project-name' or 'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL
- `source_branch` (string, required): Source branch name
- `target_branch` (string, required): Target branch name
- `title` (string, required): MR title
- `description` (string, optional): MR description
- `assignee_ids` (array, optional): Array of assignee user IDs
- `reviewer_ids` (array, optional): Array of reviewer user IDs
- `labels` (array, optional): Array of labels
- `milestone_id` (number, optional): Milestone ID
- `remove_source_branch` (boolean, optional): Auto-remove source branch after merge (default: true)
- `draft` (boolean, optional): Create as draft MR (default: false)

**Examples:**
```
Create a merge request from feature-branch to main for project namespace/project_name
```

```
Create a merge request with assignees and labels from feature-branch to main for project namespace/project_name
```

```
Create a draft merge request with description "New feature implementation" from feature-branch to develop for project namespace/project_name
```

**Response Format:**
Returns a JSON object of the created merge request containing:
- `id`: Merge request ID
- `iid`: Internal merge request ID
- `title`: MR title
- `description`: MR description
- `state`: MR state (opened, closed, merged)
- `source_branch`: Source branch name
- `target_branch`: Target branch name
- `author`: Author object with id, username, and name
- `assignees`: Array of assignee objects
- `reviewers`: Array of reviewer objects
- `labels`: Array of label names
- `milestone`: Milestone object (if assigned)
- `web_url`: Web URL to the merge request
- `draft`: Boolean indicating if this is a draft MR
- `created_at`: Creation timestamp
- `updated_at`: Last update timestamp

### get_project_description

Retrieves the description of a GitLab project.

**Parameters:**
- `project_path` (string, required): GitLab project path including all namespaces (e.g., 'namespace/project-name' or 'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL

**Examples:**
```
Get the project description for namespace/project_name
```

**Response Format:**
Returns a JSON object containing:
- `id`: Project ID
- `name`: Project name
- `path`: Project path
- `description`: Project description

### update_project_description

Updates the description of a GitLab project.

**Parameters:**
- `project_path` (string, required): GitLab project path including all namespaces (e.g., 'namespace/project-name' or 'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL
- `description` (string, required): The new description for the project

**Examples:**
```
Update the description of namespace/project_name to "A new and improved project description"
```

**Response Format:**
Returns a JSON object containing:
- `id`: Project ID
- `name`: Project name
- `path`: Project path
- `description`: Updated project description
- `topics`: Array of project topics

### get_project_topics

Retrieves the topics/tags of a GitLab project.

**Parameters:**
- `project_path` (string, required): GitLab project path including all namespaces (e.g., 'namespace/project-name' or 'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL

**Examples:**
```
Get the topics for namespace/project_name
```

**Response Format:**
Returns a JSON object containing:
- `id`: Project ID
- `name`: Project name
- `path`: Project path
- `topics`: Array of topic strings

### update_project_topics

Updates the topics/tags of a GitLab project (replaces all existing topics).

**Parameters:**
- `project_path` (string, required): GitLab project path including all namespaces (e.g., 'namespace/project-name' or 'company/department/team/project'). Run 'git remote -v' to find the full path from the repository URL
- `topics` (array, required): Array of topic strings to set for the project (replaces all existing topics)

**Examples:**
```
Update the topics of namespace/project_name to ["golang", "mcp", "gitlab", "api"]
```

```
Remove all topics from namespace/project_name by setting an empty array []
```

**Response Format:**
Returns a JSON object containing:
- `id`: Project ID
- `name`: Project name
- `path`: Project path
- `description`: Project description
- `topics`: Updated array of topic strings

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

# Run unit tests
task test

# Show test coverage percentage
task coverage
```

### Running Tests

**Unit Tests:**
```bash
task test
```

**Coverage Testing:**
```bash
# Show coverage percentage
task coverage
```

The unit tests use interface abstractions and mocking to provide fast, reliable testing without external dependencies. Current test coverage is **78.8%**.

### Manual Testing

See [MANUAL_TEST.md](MANUAL_TEST.md) for step-by-step manual testing instructions.

### Automated Testing

Run the test script:
```bash
./test_working.sh
```

## Configuration

The MCP server runs as a subprocess and communicates via JSON-RPC over stdin/stdout. No additional configuration is required.

### MCP Transport Protocol

This server uses **stdio (stdin/stdout)** for communication, which is the standard and recommended approach for MCP servers:

**✅ Why stdio is optimal:**
- **Standard MCP pattern** - Most MCP servers use stdio communication
- **Simple process model** - Started by Claude Code as a subprocess  
- **No port conflicts** - No network port management needed
- **Security** - Process isolation, no network exposure
- **Resource efficiency** - Direct pipe communication with minimal overhead
- **Cross-platform compatibility** - Works everywhere Go works
- **Easy configuration** - Just specify the executable path in Claude Code

**🌐 Alternative transports (HTTP/SSE):**
While MCP supports HTTP and Server-Sent Events, these are better suited for:
- Multi-client scenarios (serving multiple agents simultaneously)
- Containerized environments where process spawning is restricted  
- Remote access across network boundaries
- Web-based MCP clients

**For GitLab integration with Claude Code, stdio provides the best user experience** with simple configuration and reliable performance.

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