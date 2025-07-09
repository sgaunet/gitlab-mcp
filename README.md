# GitLab MCP Server

A Model Context Protocol (MCP) server that provides GitLab integration tools for Claude Code.

## Features

- **Get Project ID**: Extract GitLab project ID from remote repository URLs
- Support for both SSH and HTTPS GitLab URLs
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

### get_project_id

Extracts the GitLab project ID from a repository URL.

**Parameters:**
- `remote_url` (string): GitLab repository URL (SSH or HTTPS format)

**Example:**
```
Give me the GitLab project ID for git@gitlab.com:example/example-project.git
```

**Supported URL formats:**
- SSH: `git@gitlab.com:user/repo.git`
- HTTPS: `https://gitlab.com/user/repo.git`

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