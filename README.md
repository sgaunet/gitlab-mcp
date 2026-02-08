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

A Model Context Protocol (MCP) server that provides GitLab integration tools for Claude Code. Interact with GitLab projects, issues, epics, and CI/CD pipelines directly from Claude.

## Features

- **Issue Management**: List, create, update issues, and add comments
- **Label Management**: List and manage project labels with filtering
- **Project Management**: View and update project descriptions and topics
- **Epic Management**: List and create epics (Premium/Ultimate tier)
- **CI/CD Integration**: Monitor pipelines, view job logs, and download traces
- **Direct Project Access**: Use project paths (namespace/project-name) without ID resolution
- **MCP Architecture**: Seamless integration with Claude Code via stdio communication

## Quick Start

### Prerequisites

- GitLab personal access token with `api`, `read_api`, and `write_api` scopes
- Claude Code CLI installed
- Docker (optional, for containerized deployment)

[Detailed setup instructions →](docs/SETUP.md)

### Installation

```bash
# Install via Homebrew (Recommended for macOS/Linux)
brew tap sgaunet/homebrew-tools
brew install sgaunet/tools/gitlab-mcp
```

### Configuration

```bash
# Set your GitLab token
export GITLAB_TOKEN=your_personal_access_token

# Optional: For self-hosted GitLab
export GITLAB_URI=https://your.gitlab.instance
```

### Add to Claude Code

```bash
# Apple Silicon Mac
claude mcp add gitlab-mcp -s user -- /opt/homebrew/bin/gitlab-mcp

# Intel Mac / Linux
claude mcp add gitlab-mcp -s user -- /usr/local/bin/gitlab-mcp
```

### First Usage

```
List all open issues for project myorg/myproject
```

```
Create an issue with title "Bug fix needed" for project myorg/myproject
```

```
Get the latest pipeline for myorg/myproject
```

## Documentation

- **[Setup Guide](docs/SETUP.md)** - Installation, configuration, and Claude Code integration
- **[Docker Deployment](docs/DOCKER.md)** - Container-based deployment and Docker Compose
- **[Tool Reference](docs/TOOLS.md)** - Complete documentation for all available tools
- **[Development Guide](docs/DEVELOPMENT.md)** - Contributing, testing, and development workflow
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[Contributing](docs/CONTRIBUTING.md)** - How to contribute to the project

## Available Tools

| Tool | Description |
|------|-------------|
| `list_issues` | List project and group issues with filtering |
| `create_issues` | Create new issues with labels and assignees |
| `update_issues` | Update issue title, description, state, labels |
| `add_issue_note` | Add comments to issues |
| `list_labels` | List project labels with optional filtering |
| `get_project_description` | Get project description |
| `update_project_description` | Update project description |
| `get_project_topics` | Get project topics/tags |
| `update_project_topics` | Update project topics/tags |
| `list_epics` | List epics for a group (Premium/Ultimate) |
| `create_epic` | Create epics (Premium/Ultimate) |
| `get_latest_pipeline` | Get latest CI/CD pipeline |
| `list_pipeline_jobs` | List pipeline jobs with filtering |
| `get_job_log` | Get complete job log output |
| `download_job_trace` | Download job logs to files |

[Complete tool documentation →](docs/TOOLS.md)

## License

MIT License. See [LICENSE](LICENSE) for details.

## Support

For issues, questions, or feature requests, please [create an issue](https://github.com/sgaunet/gitlab-mcp/issues) in the repository.
