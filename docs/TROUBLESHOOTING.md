# Troubleshooting Guide

[< Back to README](../README.md)

Common issues and solutions for gitlab-mcp.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Configuration Issues](#configuration-issues)
- [Connection Issues](#connection-issues)
- [Runtime Errors](#runtime-errors)
- [Docker Issues](#docker-issues)
- [MCP Integration Issues](#mcp-integration-issues)
- [Performance Issues](#performance-issues)
- [Debug Mode](#debug-mode)
- [Getting Help](#getting-help)

---

## Installation Issues

### Binary Not Found

**Problem:** `command not found: gitlab-mcp`

**Solutions:**

1. **Verify installation:**
   ```bash
   which gitlab-mcp
   ```

2. **Check PATH:**
   ```bash
   echo $PATH
   ```

3. **Reinstall:**
   ```bash
   # Homebrew
   brew reinstall sgaunet/tools/gitlab-mcp

   # Or move binary to PATH
   sudo mv gitlab-mcp /usr/local/bin/
   ```

4. **Verify binary location:**
   - Homebrew (Apple Silicon): `/opt/homebrew/bin/gitlab-mcp`
   - Homebrew (Intel): `/usr/local/bin/gitlab-mcp`
   - Manual install: Wherever you placed it

### Permission Denied

**Problem:** `permission denied: ./gitlab-mcp`

**Solutions:**

1. **Make executable:**
   ```bash
   chmod +x gitlab-mcp
   ```

2. **Check ownership:**
   ```bash
   ls -la gitlab-mcp
   ```

3. **Reinstall with proper permissions:**
   ```bash
   sudo mv gitlab-mcp /usr/local/bin/
   sudo chmod +x /usr/local/bin/gitlab-mcp
   ```

### Architecture Mismatches

**Problem:** `bad CPU type in executable` or `cannot execute binary file`

**Solutions:**

1. **Check your architecture:**
   ```bash
   uname -m
   ```
   - `x86_64` = Intel/AMD 64-bit
   - `arm64` / `aarch64` = ARM 64-bit (Apple Silicon)

2. **Download correct binary:**
   - Apple Silicon Mac: `darwin_arm64`
   - Intel Mac: `darwin_amd64`
   - Linux x86_64: `linux_amd64`
   - Linux ARM: `linux_arm64`

3. **Rebuild from source:**
   ```bash
   git clone https://github.com/sgaunet/gitlab-mcp.git
   cd gitlab-mcp
   task build
   ```

---

## Configuration Issues

### Missing GITLAB_TOKEN

**Problem:** `authentication failed` or `401 Unauthorized`

**Solutions:**

1. **Check token is set:**
   ```bash
   echo $GITLAB_TOKEN
   ```
   Should output your token. If empty:

2. **Set token temporarily:**
   ```bash
   export GITLAB_TOKEN=your_personal_access_token
   ```

3. **Set token permanently:**
   ```bash
   # Bash
   echo 'export GITLAB_TOKEN=your_token' >> ~/.bashrc
   source ~/.bashrc

   # Zsh
   echo 'export GITLAB_TOKEN=your_token' >> ~/.zshrc
   source ~/.zshrc
   ```

4. **Verify token works:**
   ```bash
   curl -H "PRIVATE-TOKEN: $GITLAB_TOKEN" https://gitlab.com/api/v4/user
   ```

### Invalid Token Scopes

**Problem:** `403 Forbidden` or `insufficient permissions`

**Solutions:**

1. **Check required scopes:**
   - ✅ `api` - Full API access
   - ✅ `read_api` - Read access
   - ✅ `write_api` - Write access

2. **Create new token:**
   - Go to: https://gitlab.com/-/user_settings/personal_access_tokens
   - Create token with all three scopes above

3. **Update environment variable:**
   ```bash
   export GITLAB_TOKEN=new_token_with_correct_scopes
   ```

### Self-Hosted GitLab Connection

**Problem:** Can't connect to self-hosted GitLab instance

**Solutions:**

1. **Set GITLAB_URI:**
   ```bash
   export GITLAB_URI=https://your.gitlab.instance
   ```

2. **Verify URI format:**
   - Must include `https://` or `http://`
   - Must end with `/` (e.g., `https://gitlab.example.com/`)

3. **Test connection:**
   ```bash
   curl -H "PRIVATE-TOKEN: $GITLAB_TOKEN" $GITLAB_URI/api/v4/user
   ```

4. **Check network access:**
   - Verify you can reach the GitLab instance
   - Check firewall rules
   - Verify SSL/TLS certificates

### Label Validation Errors

**Problem:** `The following labels do not exist`

**Solutions:**

1. **List available labels:**
   ```
   List all labels for project myorg/myproject
   ```

2. **Fix label typos:**
   - Check spelling
   - Check case sensitivity

3. **Disable validation (if needed):**
   ```bash
   export GITLAB_VALIDATE_LABELS=false
   ```

4. **Create missing labels in GitLab:**
   - Go to project → Labels
   - Add the labels you need

---

## Connection Issues

### Parse Error (-32700)

**Problem:** JSON parse error when using the server

**Solutions:**

1. **Use through Claude Code:**
   - Don't input raw JSON directly
   - Use natural language commands in Claude Code

2. **Verify MCP configuration:**
   ```bash
   cat ~/.config/claude/mcp.json
   ```

3. **Check binary path:**
   ```json
   {
     "mcpServers": {
       "gitlab-mcp": {
         "type": "stdio",
         "command": "/correct/path/to/gitlab-mcp"
       }
     }
   }
   ```

### Network Timeouts

**Problem:** Requests timeout when accessing GitLab

**Solutions:**

1. **Check network connectivity:**
   ```bash
   ping gitlab.com
   ```

2. **Test API directly:**
   ```bash
   curl -I https://gitlab.com/api/v4/user
   ```

3. **Check proxy settings:**
   ```bash
   echo $HTTP_PROXY
   echo $HTTPS_PROXY
   ```

4. **Increase timeout (if needed):**
   - Contact GitLab admin for rate limits
   - Check if your IP is blocked

---

## Runtime Errors

### Project Not Found

**Problem:** `404 Project Not Found`

**Solutions:**

1. **Verify project path:**
   ```bash
   git remote -v
   ```
   Extract path: `https://gitlab.com/myorg/myproject.git` → `myorg/myproject`

2. **Check project access:**
   - Log in to GitLab web interface
   - Verify you can see the project
   - Check project visibility settings

3. **Test with API:**
   ```bash
   curl -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
     "https://gitlab.com/api/v4/projects/myorg%2Fmyproject"
   ```

### Issue Not Found

**Problem:** `Issue IID not found`

**Solutions:**

1. **Use issue IID (internal ID), not global ID:**
   - ✅ Correct: Issue #42 → `issue_iid: 42`
   - ❌ Incorrect: Global ID 12345

2. **List issues to find IID:**
   ```
   List all issues for project myorg/myproject
   ```

3. **Check issue exists:**
   - View issue in GitLab web interface
   - Look at URL: `https://gitlab.com/myorg/myproject/-/issues/42`
   - The number after `/issues/` is the IID

### Permission Denied

**Problem:** `403 Forbidden` when performing operations

**Solutions:**

1. **Check token scopes:**
   - Needs: `api`, `read_api`, `write_api`

2. **Check project permissions:**
   - Need at least Developer role for most operations
   - Need Maintainer role for project settings
   - Need Owner role for sensitive operations

3. **Check group permissions (for epics):**
   - Epics require Premium/Ultimate tier
   - Need appropriate group permissions

### Rate Limiting

**Problem:** `429 Too Many Requests`

**Solutions:**

1. **Wait before retrying:**
   - Default rate limit: 600 requests per minute per IP
   - Wait 60 seconds before next request

2. **Check rate limit status:**
   ```bash
   curl -I -H "PRIVATE-TOKEN: $GITLAB_TOKEN" https://gitlab.com/api/v4/user
   ```
   Look for headers:
   - `RateLimit-Limit`
   - `RateLimit-Remaining`
   - `RateLimit-Reset`

3. **Reduce request frequency:**
   - Use pagination wisely
   - Cache results when possible
   - Batch operations

4. **Contact GitLab admin:**
   - Request higher rate limits if needed
   - Check if your IP is being throttled

---

## Docker Issues

### Container Startup Failures

**Problem:** Docker container exits immediately

**Solutions:**

1. **Check environment variables:**
   ```bash
   docker run --rm -i \
     -e GITLAB_TOKEN=$GITLAB_TOKEN \
     ghcr.io/sgaunet/gitlab-mcp:latest
   ```

2. **View container logs:**
   ```bash
   docker logs gitlab-mcp-server
   ```

3. **Test with explicit token:**
   ```bash
   docker run --rm -i \
     -e GITLAB_TOKEN=your_token_here \
     ghcr.io/sgaunet/gitlab-mcp:latest
   ```

### Environment Variables Not Passed

**Problem:** Container can't authenticate to GitLab

**Solutions:**

1. **Check shell variable expansion:**
   ```bash
   echo $GITLAB_TOKEN  # Should show your token
   ```

2. **Use env file:**
   ```bash
   docker run --rm -i \
     --env-file .env \
     ghcr.io/sgaunet/gitlab-mcp:latest
   ```

3. **Verify env file format:**
   ```bash
   cat .env
   # Should be:
   # GITLAB_TOKEN=your_token
   # Not:
   # export GITLAB_TOKEN=your_token
   ```

### Image Pull Failures

**Problem:** Can't pull Docker image

**Solutions:**

1. **Check image exists:**
   ```bash
   docker pull ghcr.io/sgaunet/gitlab-mcp:latest
   ```

2. **Try specific version:**
   ```bash
   docker pull ghcr.io/sgaunet/gitlab-mcp:0.9.1
   ```

3. **Check network connectivity:**
   ```bash
   curl -I https://ghcr.io
   ```

4. **Clear Docker cache:**
   ```bash
   docker system prune -a
   ```

For more Docker troubleshooting, see the [Docker Deployment Guide](DOCKER.md).

---

## MCP Integration Issues

### Tool Calls Failing

**Problem:** Claude Code can't execute GitLab tools

**Solutions:**

1. **Verify MCP server is registered:**
   ```bash
   claude mcp list
   ```

2. **Check configuration:**
   ```bash
   cat ~/.config/claude/mcp.json
   ```

3. **Restart Claude Code:**
   ```bash
   # Exit and restart
   claude
   ```

4. **Check binary permissions:**
   ```bash
   ls -la /opt/homebrew/bin/gitlab-mcp
   ```

5. **Test binary directly:**
   ```bash
   echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | gitlab-mcp
   ```

### Communication Timeouts

**Problem:** MCP tool calls timeout

**Solutions:**

1. **Check GitLab API response time:**
   ```bash
   time curl -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
     https://gitlab.com/api/v4/user
   ```

2. **Reduce request size:**
   - Use smaller limits for list operations
   - Filter results more aggressively

3. **Check network latency:**
   - Test network speed to GitLab
   - Check for proxy issues

---

## Performance Issues

### Slow List Operations

**Problem:** Listing issues/labels takes too long

**Solutions:**

1. **Use pagination:**
   ```
   List issues with limit=20 for project myorg/myproject
   ```

2. **Filter results:**
   ```
   List issues with state=opened and labels="bug" for project myorg/myproject
   ```

3. **Disable group issues (if not needed):**
   ```
   List issues for myorg/myproject with include_group_issues=false
   ```

### Large Log Downloads

**Problem:** Job log downloads are slow or fail

**Solutions:**

1. **Check log size first:**
   - View job in GitLab web interface
   - Check file size before downloading

2. **Use streaming:**
   - `get_job_log` for viewing
   - `download_job_trace` for saving

3. **Download to fast storage:**
   ```
   Download the log for job 12345 to /tmp/job.log
   ```

---

## Debug Mode

### Enable Debug Logging

The server outputs debug information to stderr:

```bash
# Run locally with debug output
export GITLAB_TOKEN=your_token
go run . 2> debug.log

# View debug logs
tail -f debug.log
```

### Test with Echo Pipe

```bash
# Test initialize
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | gitlab-mcp

# Check output for errors
```

### Check Claude Code Logs

Claude Code may log MCP errors. Check:
- Terminal output when running Claude Code
- Claude Code settings for log locations
- System logs for process errors

---

## Getting Help

If you've tried the solutions above and still have issues:

### 1. Check Existing Issues

Search the [issue tracker](https://github.com/sgaunet/gitlab-mcp/issues) for similar problems.

### 2. Create a New Issue

When creating an issue, include:

- **Environment:**
  - OS and version
  - Go version (if building from source)
  - Docker version (if using containers)
  - GitLab version (if self-hosted)

- **Configuration:**
  - How you installed (Homebrew, binary, Docker, source)
  - MCP configuration (`.mcp.json` without tokens)
  - Environment variables set (without token values)

- **Error Details:**
  - Exact error message
  - Steps to reproduce
  - Expected vs actual behavior
  - Debug logs (if available)

- **What You've Tried:**
  - Solutions attempted from this guide
  - Any workarounds that partially worked

### 3. Community Support

- **GitHub Discussions**: Ask questions and share tips
- **Documentation**: Check all docs in [docs/](.)
- **Examples**: Review [MANUAL_TEST.md](../MANUAL_TEST.md)

---

## Next Steps

- [Setup guide →](SETUP.md)
- [Tool reference →](TOOLS.md)
- [Development guide →](DEVELOPMENT.md)
- [Docker deployment →](DOCKER.md)
