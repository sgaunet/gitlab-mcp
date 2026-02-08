# Docker Deployment Guide

[< Back to README](../README.md)

Comprehensive guide for deploying the GitLab MCP Server using Docker containers.

## Table of Contents

- [Why Docker?](#why-docker)
- [Using Pre-built Images](#using-pre-built-images)
- [Claude Code Integration](#claude-code-integration)
- [Docker Compose](#docker-compose)
- [Environment Variables](#environment-variables)
- [Building Custom Images](#building-custom-images)
- [Troubleshooting](#troubleshooting)

## Why Docker?

Docker deployment is ideal for:
- **Containerized development environments** - Consistent runtime across teams
- **CI/CD scenarios** - Integrate GitLab MCP into automated workflows
- **Dependency isolation** - No Go installation required on host
- **Easy deployment** - Single command to run or update
- **Multi-platform support** - Pre-built images for amd64 and arm64

## Using Pre-built Images

### Available Images

Pre-built Docker images are published to GitHub Container Registry:

```bash
# Latest stable release
ghcr.io/sgaunet/gitlab-mcp:latest

# Specific version
ghcr.io/sgaunet/gitlab-mcp:0.9.1
```

### Pull the Image

```bash
docker pull ghcr.io/sgaunet/gitlab-mcp:latest
```

### Basic Usage

Run with GitLab.com:

```bash
docker run --rm -i \
  -e GITLAB_TOKEN=your_personal_access_token \
  ghcr.io/sgaunet/gitlab-mcp:latest
```

### Self-Hosted GitLab

For self-hosted GitLab instances:

```bash
docker run --rm -i \
  -e GITLAB_TOKEN=your_personal_access_token \
  -e GITLAB_URI=https://your.gitlab.instance \
  ghcr.io/sgaunet/gitlab-mcp:latest
```

### Disable Label Validation

```bash
docker run --rm -i \
  -e GITLAB_TOKEN=your_personal_access_token \
  -e GITLAB_VALIDATE_LABELS=false \
  ghcr.io/sgaunet/gitlab-mcp:latest
```

### Docker Run Flags Explained

- `--rm` - Automatically remove container when it stops
- `-i` - Keep stdin open for MCP communication
- `-e` - Set environment variables

## Claude Code Integration

### Method 1: Using docker run

Add the Docker-based MCP server to Claude Code:

```bash
# Using environment variable from shell
claude mcp add gitlab-mcp-docker -s user -- \
  docker run --rm -i \
  -e GITLAB_TOKEN=${GITLAB_TOKEN} \
  ghcr.io/sgaunet/gitlab-mcp:latest

# With explicit token (not recommended)
claude mcp add gitlab-mcp-docker -s user -- \
  docker run --rm -i \
  -e GITLAB_TOKEN=your_token_here \
  ghcr.io/sgaunet/gitlab-mcp:latest
```

### Method 2: Using .mcp.json

Configure directly in Claude Code's MCP configuration file:

**Location:** `~/.config/claude/mcp.json` (Linux/macOS) or `%APPDATA%\claude\mcp.json` (Windows)

```json
{
  "mcpServers": {
    "gitlab-mcp-docker": {
      "type": "stdio",
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-e",
        "GITLAB_TOKEN=your_personal_access_token",
        "--name",
        "gitlab-mcp-server",
        "ghcr.io/sgaunet/gitlab-mcp:latest"
      ]
    }
  }
}
```

### Method 3: Using Environment File

For better security, use an environment file:

**Create `.env` file:**
```bash
GITLAB_TOKEN=your_personal_access_token
GITLAB_URI=https://gitlab.com/
GITLAB_VALIDATE_LABELS=true
```

**Add to .gitignore:**
```bash
echo ".env" >> .gitignore
```

**Update .mcp.json:**
```json
{
  "mcpServers": {
    "gitlab-mcp-docker": {
      "type": "stdio",
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "--env-file",
        "/absolute/path/to/.env",
        "--name",
        "gitlab-mcp-server",
        "ghcr.io/sgaunet/gitlab-mcp:latest"
      ]
    }
  }
}
```

**Important:** Use absolute paths for `--env-file` as Docker requires them.

### Security Best Practices

✅ **DO:**
- Use environment variables from your shell profile
- Store tokens in `.env` files (never commit them)
- Use secret management tools (Vault, AWS Secrets Manager, etc.)
- Set appropriate file permissions on `.env` files (`chmod 600 .env`)

❌ **DON'T:**
- Hardcode tokens in configuration files
- Commit `.env` files to version control
- Share tokens in plaintext
- Use overly permissive token scopes

## Docker Compose

For managing the container lifecycle with Docker Compose:

### Create docker-compose.yml

```yaml
services:
  gitlab-mcp:
    image: ghcr.io/sgaunet/gitlab-mcp:latest
    container_name: gitlab-mcp-server
    stdin_open: true
    tty: true
    env_file: .env
    restart: unless-stopped
```

### Create .env File

```bash
# GitLab MCP Server Environment Variables

# Required: GitLab Personal Access Token
# Scopes needed: api, read_api, write_api
# Create at: https://gitlab.com/-/user_settings/personal_access_tokens
GITLAB_TOKEN=your_personal_access_token_here

# Optional: GitLab Instance URI (defaults to https://gitlab.com/)
# For self-hosted instances, set to your GitLab URL
GITLAB_URI=https://gitlab.com/

# Optional: Label Validation (defaults to true)
# true: Validates labels exist before creating issues
# false: Allows non-existent labels (GitLab's default behavior)
GITLAB_VALIDATE_LABELS=true
```

### Add to .gitignore

```bash
echo ".env" >> .gitignore
```

### Manage the Container

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down

# Restart the service
docker-compose restart

# Pull latest image
docker-compose pull
docker-compose up -d
```

### Integrate with Claude Code

While Docker Compose is useful for managing containers, Claude Code expects stdio communication. For Claude Code integration, use the `.mcp.json` configuration:

```json
{
  "mcpServers": {
    "gitlab-mcp-docker": {
      "type": "stdio",
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "--env-file",
        ".env",
        "--name",
        "gitlab-mcp-server",
        "ghcr.io/sgaunet/gitlab-mcp:latest"
      ]
    }
  }
}
```

## Environment Variables

### Complete Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GITLAB_TOKEN` | Yes | - | GitLab personal access token with `api`, `read_api`, `write_api` scopes |
| `GITLAB_URI` | No | `https://gitlab.com/` | GitLab instance URI (for self-hosted instances) |
| `GITLAB_VALIDATE_LABELS` | No | `true` | Enable/disable label validation for issue creation |

### Setting Multiple Variables

```bash
docker run --rm -i \
  -e GITLAB_TOKEN=your_token \
  -e GITLAB_URI=https://your.gitlab.instance \
  -e GITLAB_VALIDATE_LABELS=false \
  ghcr.io/sgaunet/gitlab-mcp:latest
```

## Building Custom Images

### Using Dockerfile

If you need to build a custom image:

1. **Clone the repository:**
   ```bash
   git clone https://github.com/sgaunet/gitlab-mcp.git
   cd gitlab-mcp
   ```

2. **Build the image:**
   ```bash
   docker build -t my-gitlab-mcp:latest .
   ```

3. **Run your custom image:**
   ```bash
   docker run --rm -i \
     -e GITLAB_TOKEN=your_token \
     my-gitlab-mcp:latest
   ```

### Multi-Architecture Builds

Build for multiple architectures using buildx:

```bash
# Create a new builder
docker buildx create --use

# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t my-gitlab-mcp:latest \
  --push .
```

## Troubleshooting

### Container Won't Start

**Problem:** Container exits immediately

**Solutions:**
1. Check token is set:
   ```bash
   docker run --rm -i \
     -e GITLAB_TOKEN=your_token \
     ghcr.io/sgaunet/gitlab-mcp:latest
   ```

2. Verify environment file exists:
   ```bash
   ls -la .env
   ```

3. Check Docker logs:
   ```bash
   docker logs gitlab-mcp-server
   ```

### Environment Variables Not Passed

**Problem:** Server can't connect to GitLab

**Solutions:**
1. Verify variables are set:
   ```bash
   docker run --rm -i \
     -e GITLAB_TOKEN=${GITLAB_TOKEN} \
     ghcr.io/sgaunet/gitlab-mcp:latest
   ```

2. Check shell expansion:
   ```bash
   echo $GITLAB_TOKEN  # Should output your token
   ```

3. Use explicit values for testing:
   ```bash
   docker run --rm -i \
     -e GITLAB_TOKEN=glpat-your_token \
     ghcr.io/sgaunet/gitlab-mcp:latest
   ```

### Permission Errors

**Problem:** Docker can't read .env file

**Solutions:**
1. Fix file permissions:
   ```bash
   chmod 644 .env
   ```

2. Use absolute path:
   ```bash
   docker run --rm -i \
     --env-file /absolute/path/to/.env \
     ghcr.io/sgaunet/gitlab-mcp:latest
   ```

### Image Pull Failures

**Problem:** Can't pull image from registry

**Solutions:**
1. Check network connectivity
2. Verify image tag exists:
   ```bash
   docker pull ghcr.io/sgaunet/gitlab-mcp:latest
   ```
3. Try specific version:
   ```bash
   docker pull ghcr.io/sgaunet/gitlab-mcp:0.9.1
   ```

### Claude Code Integration Issues

**Problem:** Claude Code can't communicate with Docker container

**Solutions:**
1. Ensure `-i` flag is present (keeps stdin open)
2. Ensure `--rm` flag is present (cleans up containers)
3. Check MCP configuration syntax:
   ```json
   {
     "type": "stdio",
     "command": "docker",
     "args": ["run", "--rm", "-i", ...]
   }
   ```

For more troubleshooting help, see the [Troubleshooting Guide](TROUBLESHOOTING.md).

## Next Steps

- [Learn about all available tools →](TOOLS.md)
- [View complete setup guide →](SETUP.md)
- [Set up development environment →](DEVELOPMENT.md)
