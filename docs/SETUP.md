# Setup Guide

[< Back to README](../README.md)

Complete installation and configuration guide for the GitLab MCP Server.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Option 1: Homebrew (Recommended)](#option-1-homebrew-recommended)
  - [Option 2: GitHub Releases](#option-2-github-releases)
  - [Option 3: Build from Source](#option-3-build-from-source)
  - [Option 4: Docker](#option-4-docker)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Setting Variables Permanently](#setting-variables-permanently)
  - [GitLab Token Creation](#gitlab-token-creation)
  - [MCP Transport Protocol](#mcp-transport-protocol)
- [Adding to Claude Code](#adding-to-claude-code)
- [Verification](#verification)

## Prerequisites

- **Go 1.21 or later** (required for building from source)
- **[Task](https://taskfile.dev/)** (optional, for build automation)
- **Claude Code CLI** or any MCP-compatible client
- **GitLab Access**: Access to GitLab repositories (GitLab.com or self-hosted)
- **GitLab Personal Access Token**: Token with appropriate scopes

### Required Token Scopes

Your GitLab personal access token must have these scopes:
- `api` - Full API access (includes read and write)
- `read_api` - Read-only API access
- `write_api` - Write access to API

## Installation

### Option 1: Homebrew (Recommended)

Best for macOS and Linux users. Provides automatic updates and easy management.

```bash
# Add the tap
brew tap sgaunet/homebrew-tools

# Install the MCP server
brew install sgaunet/tools/gitlab-mcp

# Verify installation
gitlab-mcp --version
```

**Installation Locations:**
- **Apple Silicon Mac**: `/opt/homebrew/bin/gitlab-mcp`
- **Intel Mac**: `/usr/local/bin/gitlab-mcp`
- **Linux**: `/home/linuxbrew/.linuxbrew/bin/gitlab-mcp`

### Option 2: GitHub Releases

Download pre-built binaries for your platform.

1. **Visit the releases page:**

   Go to [github.com/sgaunet/gitlab-mcp/releases/latest](https://github.com/sgaunet/gitlab-mcp/releases/latest)

2. **Download the appropriate binary:**

   - **macOS Intel**: `gitlab-mcp_VERSION_darwin_amd64`
   - **macOS Apple Silicon**: `gitlab-mcp_VERSION_darwin_arm64`
   - **Linux x86_64**: `gitlab-mcp_VERSION_linux_amd64`
   - **Linux ARM64**: `gitlab-mcp_VERSION_linux_arm64`
   - **Windows**: `gitlab-mcp_VERSION_windows_amd64.exe`

3. **Make it executable (macOS/Linux):**
   ```bash
   chmod +x gitlab-mcp_*
   ```

4. **Move to a location in your PATH:**
   ```bash
   # macOS/Linux
   sudo mv gitlab-mcp_* /usr/local/bin/gitlab-mcp

   # Verify
   gitlab-mcp --version
   ```

5. **Windows Installation:**
   ```powershell
   # Move to a directory in your PATH
   move gitlab-mcp_*.exe C:\Windows\System32\gitlab-mcp.exe

   # Or add to a custom directory and update PATH
   ```

### Option 3: Build from Source

For developers or users needing the latest changes.

1. **Clone the repository:**
   ```bash
   git clone https://github.com/sgaunet/gitlab-mcp.git
   cd gitlab-mcp
   ```

2. **Build using Task (recommended):**
   ```bash
   task build
   ```

   Or build manually:
   ```bash
   go build -o gitlab-mcp
   ```

3. **Install to your PATH:**
   ```bash
   sudo mv gitlab-mcp /usr/local/bin/
   ```

4. **Verify installation:**
   ```bash
   gitlab-mcp --version
   ```

### Option 4: Docker

See the [Docker Deployment Guide](DOCKER.md) for containerized deployment options.

## Configuration

### Environment Variables

The MCP server requires configuration via environment variables:

#### Required Variables

**`GITLAB_TOKEN`** (required)
- Your GitLab personal access token
- Must have `api`, `read_api`, and `write_api` scopes
- Never commit this token to version control

```bash
export GITLAB_TOKEN=your_personal_access_token
```

#### Optional Variables

**`GITLAB_URI`** (optional)
- GitLab instance URI
- Default: `https://gitlab.com/`
- Set this for self-hosted GitLab instances

```bash
export GITLAB_URI=https://your.gitlab.instance
```

**`GITLAB_VALIDATE_LABELS`** (optional)
- Enable/disable label validation for issue creation
- Default: `true`
- Values:
  - `true`: Validates labels exist before creating issues (prevents typos)
  - `false`: Allows non-existent labels (GitLab's default behavior)

```bash
export GITLAB_VALIDATE_LABELS=true   # Default
export GITLAB_VALIDATE_LABELS=false  # Disable validation
```

### Setting Variables Permanently

To avoid setting environment variables every session, add them to your shell profile:

**Bash:**
```bash
# Add to ~/.bashrc or ~/.bash_profile
echo 'export GITLAB_TOKEN=your_personal_access_token' >> ~/.bashrc
echo 'export GITLAB_URI=https://gitlab.com/' >> ~/.bashrc
source ~/.bashrc
```

**Zsh:**
```bash
# Add to ~/.zshrc
echo 'export GITLAB_TOKEN=your_personal_access_token' >> ~/.zshrc
echo 'export GITLAB_URI=https://gitlab.com/' >> ~/.zshrc
source ~/.zshrc
```

**Fish:**
```fish
# Add to ~/.config/fish/config.fish
echo 'set -gx GITLAB_TOKEN your_personal_access_token' >> ~/.config/fish/config.fish
echo 'set -gx GITLAB_URI https://gitlab.com/' >> ~/.config/fish/config.fish
source ~/.config/fish/config.fish
```

**Important:** Setting variables in shell profiles ensures they're available when Claude Code starts the MCP server as a subprocess.

### GitLab Token Creation

1. **Sign in to GitLab**
   - GitLab.com: [gitlab.com](https://gitlab.com)
   - Self-hosted: Your GitLab instance URL

2. **Navigate to Personal Access Tokens**
   - Click your avatar (top right) → Settings → Access Tokens
   - Or go directly to: `https://gitlab.com/-/user_settings/personal_access_tokens`

3. **Create a new token**
   - Name: `claude-code-mcp` (or any descriptive name)
   - Expiration date: Set according to your security policy
   - Scopes: Select:
     - ✅ `api` - Full API access
     - ✅ `read_api` - Read access
     - ✅ `write_api` - Write access

4. **Copy the token**
   - Save it immediately - you won't be able to see it again
   - Store securely (password manager recommended)

5. **Set the environment variable**
   ```bash
   export GITLAB_TOKEN=your_copied_token
   ```

### MCP Transport Protocol

This server uses **stdio (stdin/stdout)** for communication, which is the standard MCP transport:

#### Why stdio?

✅ **Standard MCP pattern** - Most MCP servers use stdio communication
✅ **Simple process model** - Started by Claude Code as a subprocess
✅ **No port conflicts** - No network port management needed
✅ **Security** - Process isolation, no network exposure
✅ **Resource efficiency** - Direct pipe communication with minimal overhead
✅ **Cross-platform compatibility** - Works everywhere Go works
✅ **Easy configuration** - Just specify the executable path

#### Alternative Transports

While MCP supports HTTP and Server-Sent Events (SSE), these are better suited for:
- Multi-client scenarios (serving multiple agents simultaneously)
- Containerized environments where process spawning is restricted
- Remote access across network boundaries
- Web-based MCP clients

**For GitLab integration with Claude Code, stdio provides the best user experience.**

## Adding to Claude Code

After installation and configuration, register the MCP server with Claude Code.

### Command-Line Registration

```bash
# Apple Silicon Mac (Homebrew)
claude mcp add gitlab-mcp -s user -- /opt/homebrew/bin/gitlab-mcp

# Intel Mac or Linux (Homebrew)
claude mcp add gitlab-mcp -s user -- /usr/local/bin/gitlab-mcp

# Custom installation path
claude mcp add gitlab-mcp -s user -- /path/to/gitlab-mcp

# Docker-based (see Docker guide for details)
claude mcp add gitlab-mcp-docker -s user -- \
  docker run --rm -i \
  -e GITLAB_TOKEN=${GITLAB_TOKEN} \
  ghcr.io/sgaunet/gitlab-mcp:latest
```

### Manual Configuration

Alternatively, edit your MCP configuration file directly:

**Location:** `~/.config/claude/mcp.json` (Linux/macOS) or `%APPDATA%\claude\mcp.json` (Windows)

```json
{
  "mcpServers": {
    "gitlab-mcp": {
      "type": "stdio",
      "command": "/opt/homebrew/bin/gitlab-mcp"
    }
  }
}
```

For Docker-based deployment:
```json
{
  "mcpServers": {
    "gitlab-mcp": {
      "type": "stdio",
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-e",
        "GITLAB_TOKEN=your_token",
        "ghcr.io/sgaunet/gitlab-mcp:latest"
      ]
    }
  }
}
```

**Security Note:** Avoid hardcoding tokens in configuration files. Use environment variables or secret management tools.

## Verification

### Test the Installation

1. **Start Claude Code:**
   ```bash
   claude
   ```

2. **Test GitLab connectivity:**
   ```
   List all open issues for project myorg/myproject
   ```

3. **Verify tool availability:**
   Claude should recognize and execute GitLab MCP tools. Look for responses that include issue data from your GitLab projects.

### Troubleshooting Connection Issues

If tools aren't working:

1. **Check token is set:**
   ```bash
   echo $GITLAB_TOKEN
   ```

2. **Verify binary location:**
   ```bash
   which gitlab-mcp
   ```

3. **Test binary directly:**
   ```bash
   echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | gitlab-mcp
   ```

4. **Check Claude Code logs:**
   Look for MCP server errors in Claude Code output.

5. **Verify token scopes:**
   Ensure your token has `api`, `read_api`, and `write_api` scopes.

For more troubleshooting help, see the [Troubleshooting Guide](TROUBLESHOOTING.md).

## Next Steps

- [Learn about all available tools →](TOOLS.md)
- [Deploy with Docker →](DOCKER.md)
- [Set up development environment →](DEVELOPMENT.md)
