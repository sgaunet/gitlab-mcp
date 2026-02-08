# Tool Reference

[< Back to README](../README.md)

Complete reference for all GitLab MCP tools available in Claude Code.

## Table of Contents

- [Issue Management](#issue-management)
  - [list_issues](#list_issues)
  - [create_issues](#create_issues)
  - [update_issues](#update_issues)
  - [add_issue_note](#add_issue_note)
- [Label Management](#label-management)
  - [list_labels](#list_labels)
- [Project Management](#project-management)
  - [get_project_description](#get_project_description)
  - [update_project_description](#update_project_description)
  - [get_project_topics](#get_project_topics)
  - [update_project_topics](#update_project_topics)
- [Epic Management (Premium/Ultimate)](#epic-management-premiumultimate)
  - [list_epics](#list_epics)
  - [create_epic](#create_epic)
  - [add_issue_to_epic](#add_issue_to_epic)
- [CI/CD Pipeline Management](#cicd-pipeline-management)
  - [get_latest_pipeline](#get_latest_pipeline)
  - [list_pipeline_jobs](#list_pipeline_jobs)
  - [get_job_log](#get_job_log)
  - [download_job_trace](#download_job_trace)
- [Common Patterns](#common-patterns)
- [Error Handling](#error-handling)

---

## Issue Management

### list_issues

Lists issues for a GitLab project using the project path. By default, includes issues from both the project and its parent group(s) for comprehensive visibility.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces (e.g., `namespace/project-name`) |
| `state` | string | No | `opened` | Filter by issue state: `opened`, `closed`, or `all` |
| `labels` | string | No | - | Comma-separated list of labels to filter by |
| `limit` | number | No | `100` | Maximum number of issues to return (max: 100) |
| `include_group_issues` | boolean | No | `true` | Include issues from parent group(s). Set to `false` for project-only issues |

#### Examples

**List all open issues:**
```
List all open issues for project myorg/myproject
```

**List all issues (open and closed):**
```
List all issues for project myorg/myproject with state=all
```

**List issues with specific labels:**
```
List issues with labels "bug,critical" for project myorg/myproject
```

**List only project-level issues (exclude group issues):**
```
List issues for myorg/myproject with include_group_issues=false
```

**List with pagination:**
```
List issues with limit=50 for project myorg/myproject
```

#### Response Format

Returns a JSON array of issue objects:

```json
[
  {
    "id": 12345,
    "iid": 42,
    "title": "Bug in authentication",
    "description": "Users cannot log in with SSO",
    "state": "opened",
    "labels": ["bug", "priority-high"],
    "assignees": [
      {
        "id": 123,
        "username": "johndoe",
        "name": "John Doe"
      }
    ],
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-16T14:20:00Z"
  }
]
```

#### Notes

- **Group Issues Integration**: By default, returns issues from both the project and parent group(s)
- **Deduplication**: Issues are automatically deduplicated by `ProjectID`
- **Graceful Fallback**: If group fetching fails, returns project-only issues
- **Performance**: Group issue fetching may increase response time for large groups

---

### create_issues

Creates a new issue for a GitLab project.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |
| `title` | string | Yes | - | Issue title |
| `description` | string | No | - | Issue description (supports Markdown) |
| `labels` | array | No | `[]` | Array of label names to assign |
| `assignees` | array | No | `[]` | Array of user IDs to assign |

#### Examples

**Basic issue creation:**
```
Create an issue with title "Bug fix needed" for project myorg/myproject
```

**Issue with description:**
```
Create an issue with title "Feature request" and description "Add dark mode support" for project myorg/myproject
```

**Issue with labels:**
```
Create an issue with title "Performance issue", description "API response time is slow", and labels ["bug", "performance"] for project myorg/myproject
```

**Issue with labels and assignees:**
```
Create an issue with title "Security vulnerability", labels ["security", "critical"], and assignees [123, 456] for project myorg/myproject
```

#### Response Format

Returns a JSON object of the created issue with the same structure as `list_issues`.

#### Label Validation

By default, the server validates that labels exist in the project before creating issues:

**Validation Enabled (default):**
- Prevents typos and invalid label names
- Returns helpful error message listing available labels
- Set `GITLAB_VALIDATE_LABELS=true` environment variable

**Example validation error:**
```
The following labels do not exist in project 'myorg/myproject':
- 'nonexistent-label'
- 'typo-label'

Available labels in this project:
- bug, enhancement, documentation, priority-high, priority-medium, priority-low

To disable label validation, set GITLAB_VALIDATE_LABELS=false
```

**Validation Disabled:**
- Allows non-existent labels (GitLab's default behavior)
- Labels are created automatically if they don't exist
- Set `GITLAB_VALIDATE_LABELS=false` environment variable

#### Notes

- Use `list_labels` to see available labels before creating issues
- Issue description supports full Markdown syntax
- Assignee IDs can be found using GitLab's API or web interface

---

### update_issues

Updates an existing issue for a GitLab project.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |
| `issue_iid` | number | Yes | - | Issue internal ID (IID) to update |
| `title` | string | No | - | Updated issue title |
| `description` | string | No | - | Updated issue description |
| `state` | string | No | - | Issue state: `opened` or `closed` |
| `labels` | array | No | - | Array of label names (replaces all existing labels) |
| `assignees` | array | No | - | Array of user IDs (replaces all existing assignees) |

#### Examples

**Update title:**
```
Update issue #5 title to "Fixed: Authentication bug" for project myorg/myproject
```

**Close an issue:**
```
Close issue #10 for project myorg/myproject
```

**Update description and state:**
```
Update issue #12 with description "Fixed in commit abc123" and state "closed" for project myorg/myproject
```

**Update labels:**
```
Update issue #15 with labels ["bug", "resolved"] for project myorg/myproject
```

**Update multiple fields:**
```
Update issue #20 with title "Resolved: Performance issue", state "closed", and labels ["performance", "resolved"] for project myorg/myproject
```

#### Response Format

Returns a JSON object of the updated issue with the same structure as `list_issues`.

#### Notes

- At least one optional parameter must be provided
- Labels and assignees are replaced entirely (not appended)
- Use issue IID (internal ID), not the global issue ID

---

### add_issue_note

Adds a note/comment to an existing issue for a GitLab project.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |
| `issue_iid` | number | Yes | - | Issue internal ID (IID) to add note to |
| `body` | string | Yes | - | Note/comment body text (supports Markdown) |

#### Examples

**Simple comment:**
```
Add a comment "This looks good to me!" to issue #5 for project myorg/myproject
```

**Detailed note:**
```
Add a note "Fixed in commit abc123. Please test and verify." to issue #12 for project myorg/myproject
```

**Markdown formatting:**
```
Add a comment with markdown "## Test Results\n- ✅ Unit tests pass\n- ✅ Integration tests pass" to issue #20 for project myorg/myproject
```

#### Response Format

Returns a JSON object of the created note:

```json
{
  "id": 789,
  "body": "This looks good to me!",
  "author": {
    "id": 123,
    "username": "johndoe",
    "name": "John Doe"
  },
  "created_at": "2024-01-16T15:30:00Z",
  "updated_at": "2024-01-16T15:30:00Z",
  "system": false,
  "noteable": {
    "id": 12345,
    "iid": 42,
    "title": "Bug in authentication"
  }
}
```

#### Notes

- Supports full Markdown syntax including code blocks, lists, and formatting
- Use issue IID (internal ID), not the global issue ID
- Notes are visible to all users with access to the issue

---

## Label Management

### list_labels

Lists labels for a GitLab project with optional filtering.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |
| `with_counts` | boolean | No | `false` | Include issue and merge request counts |
| `include_ancestor_groups` | boolean | No | `false` | Include labels from ancestor groups |
| `search` | string | No | - | Filter labels by search keyword |
| `limit` | number | No | `100` | Maximum number of labels to return (max: 100) |

#### Examples

**List all labels:**
```
List all labels for project myorg/myproject
```

**List labels with counts:**
```
List labels with counts for project myorg/myproject
```

**Search for specific labels:**
```
Search for labels containing "bug" in project myorg/myproject
```

**Include group labels:**
```
List labels with include_ancestor_groups=true for project myorg/myproject
```

#### Response Format

Returns a JSON array of label objects:

```json
[
  {
    "id": 1,
    "name": "bug",
    "color": "#FF0000",
    "text_color": "#FFFFFF",
    "description": "Something isn't working",
    "open_issues_count": 5,
    "closed_issues_count": 12
  }
]
```

#### Notes

- Use this before creating issues to see available labels
- Ancestor group labels are useful for organization-wide label standards
- Counts are only included when `with_counts=true`

---

## Project Management

### get_project_description

Retrieves the description of a GitLab project.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |

#### Examples

```
Get the project description for myorg/myproject
```

```
Get the description of sgaunet/gitlab-mcp
```

#### Response Format

```json
{
  "id": 12345,
  "name": "gitlab-mcp",
  "path": "gitlab-mcp",
  "description": "A Model Context Protocol server for GitLab integration"
}
```

---

### update_project_description

Updates the description of a GitLab project.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |
| `description` | string | Yes | - | The new description for the project |

#### Examples

```
Update the description of myorg/myproject to "A comprehensive GitLab integration tool for MCP"
```

```
Update the description of sgaunet/gitlab-mcp to "GitLab MCP Server - Integrate GitLab with Claude Code"
```

#### Response Format

```json
{
  "id": 12345,
  "name": "gitlab-mcp",
  "path": "gitlab-mcp",
  "description": "GitLab MCP Server - Integrate GitLab with Claude Code",
  "topics": ["golang", "gitlab", "mcp"]
}
```

---

### get_project_topics

Retrieves the topics/tags of a GitLab project.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |

#### Examples

```
Get the topics for myorg/myproject
```

#### Response Format

```json
{
  "id": 12345,
  "name": "gitlab-mcp",
  "path": "gitlab-mcp",
  "topics": ["golang", "gitlab", "mcp", "api", "automation"]
}
```

---

### update_project_topics

Updates the topics/tags of a GitLab project (replaces all existing topics).

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |
| `topics` | array | Yes | - | Array of topic strings (replaces all existing topics) |

#### Examples

**Set topics:**
```
Update the topics of myorg/myproject to ["golang", "mcp", "gitlab", "api"]
```

**Remove all topics:**
```
Remove all topics from myorg/myproject by setting topics to []
```

#### Response Format

```json
{
  "id": 12345,
  "name": "gitlab-mcp",
  "path": "gitlab-mcp",
  "description": "GitLab MCP Server",
  "topics": ["golang", "mcp", "gitlab", "api"]
}
```

#### Notes

- Topics are replaced entirely (not appended)
- Use empty array `[]` to remove all topics
- Topics improve project discoverability in GitLab

---

## Epic Management (Premium/Ultimate)

**Note:** Epics require GitLab Premium or Ultimate tier. Free/Starter tier instances will return an error.

### list_epics

Lists epics for a GitLab group.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `group_path` | string | Yes | - | GitLab group path (e.g., `myorg` or `parent/subgroup`) |
| `state` | string | No | `opened` | Filter by epic state: `opened`, `closed`, or `all` |
| `limit` | number | No | `100` | Maximum number of epics to return (max: 100) |

#### Examples

```
List epics for the myorg group
```

```
List all epics (open and closed) for the myorg/platform group
```

```
List closed epics for group myorg
```

#### Response Format

Returns a JSON array of epic objects with title, description, state, labels, and dates.

---

### create_epic

Creates a new epic in a GitLab group.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `group_path` | string | Yes | - | GitLab group path |
| `title` | string | Yes | - | Epic title |
| `description` | string | No | - | Epic description |
| `labels` | array | No | `[]` | Array of label names |
| `start_date` | string | No | - | Start date in YYYY-MM-DD format |
| `due_date` | string | No | - | Due date in YYYY-MM-DD format |
| `confidential` | boolean | No | `false` | Whether epic is confidential |

#### Examples

**Basic epic:**
```
Create an epic in myorg group with title "Q1 2024 Launch"
```

**Epic with details:**
```
Create an epic in myorg/platform group with title "Authentication Redesign", description "Modernize auth with OAuth2 and JWT", labels ["security", "high-priority"], start date "2024-03-01", due date "2024-06-30", and make it confidential
```

---

### add_issue_to_epic

Attaches an issue to an epic (Premium/Ultimate tier).

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `group_path` | string | Yes | - | GitLab group path containing the epic |
| `epic_iid` | number | Yes | - | Epic internal ID (IID) |
| `project_path` | string | Yes | - | GitLab project path containing the issue |
| `issue_iid` | number | Yes | - | Issue internal ID (IID) |

#### Examples

```
Add issue #42 from project myorg/backend to epic #5 in group myorg
```

#### Notes

- **Deprecation Notice**: The Epics API is deprecated and will be removed in GitLab API v5
- Consider using Work Items API for future implementations

---

## CI/CD Pipeline Management

### get_latest_pipeline

Gets the latest pipeline for a GitLab project with optional ref (branch/tag) filtering.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |
| `ref` | string | No | - | Filter by branch or tag name (e.g., `main`, `develop`, `v1.0.0`) |

#### Examples

**Get latest pipeline:**
```
Get the latest pipeline for myorg/myproject
```

**Get latest pipeline for specific branch:**
```
Get the latest pipeline for main branch in myorg/myproject
```

```
Get the latest pipeline for develop branch in myorg/myproject
```

#### Response Format

Returns pipeline metadata including ID, status, ref, sha, and web URL.

---

### list_pipeline_jobs

Lists all jobs for a GitLab pipeline with filtering options.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |
| `pipeline_id` | number | No | (latest) | Specific pipeline ID. If not provided, uses the latest pipeline |
| `ref` | string | No | - | Branch or tag name for finding the latest pipeline |
| `scope` | array | No | - | Filter by status: `created`, `pending`, `running`, `success`, `failed`, `canceled`, `skipped`, `manual`, `scheduled` |
| `stage` | string | No | - | Filter by stage name (e.g., `build`, `test`, `deploy`) |

#### Examples

**List all jobs for latest pipeline:**
```
List all jobs for the latest pipeline in myorg/myproject
```

**List jobs for specific pipeline:**
```
List all jobs for pipeline ID 12345 in myorg/myproject
```

**List failed jobs:**
```
List only failed jobs for the latest pipeline on main branch in myorg/myproject
```

**List jobs by stage:**
```
List failed jobs in the test stage for pipeline ID 12345 in myorg/myproject
```

#### Response Format

Returns a JSON array of job objects with ID, name, status, stage, and duration.

---

### get_job_log

Retrieves complete log output for a specific CI/CD job with job metadata.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |
| `job_id` | number | Yes | - | Job ID to retrieve logs for |
| `pipeline_id` | number | No | (latest) | Specific pipeline ID for context |
| `ref` | string | No | - | Branch or tag name for finding the latest pipeline |

#### Examples

**Get log for specific job:**
```
Get the log for job 12345 in pipeline 999 in myorg/myproject
```

**Get log from latest pipeline:**
```
Get the log for job 54321 from the latest pipeline in myorg/myproject
```

**Get log from specific branch:**
```
Get the log for job 11111 from the latest pipeline on develop branch in myorg/myproject
```

#### Response Format

Returns job metadata and complete log output as text.

#### Notes

- Useful for debugging job failures
- Logs can be very large - consider using `download_job_trace` for archiving

---

### download_job_trace

Downloads CI/CD job logs to local files for offline analysis and archiving.

#### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `project_path` | string | Yes | - | GitLab project path including all namespaces |
| `job_id` | number | Yes | - | Job ID to download trace for |
| `output_path` | string | No | `./job_<id>_trace.log` | File path to save trace. Parent directories created if needed |
| `pipeline_id` | number | No | (latest) | Specific pipeline ID for context |
| `ref` | string | No | - | Branch or tag name for finding the latest pipeline |

#### Examples

**Download to default location:**
```
Download the log for job 12345 to a file
```

**Download to specific path:**
```
Download the log for job 54321 from the latest pipeline to ./logs/job_54321.log
```

**Download from specific branch:**
```
Download the log for job 11111 from the develop branch to /tmp/build.log
```

#### Response Format

Returns job metadata and file information including the saved path.

#### Notes

- Parent directories are created automatically if they don't exist
- Useful for archiving logs, offline analysis, or CI/CD workflows
- Files can be large for long-running jobs

---

## Common Patterns

### Working with Project Paths

All tools use GitLab project paths in the format `namespace/project-name`:

```
myorg/myproject
company/department/team/project
username/personal-project
```

**Finding your project path:**
1. Run `git remote -v` in your repository
2. Extract the path from the URL:
   - HTTPS: `https://gitlab.com/myorg/myproject.git` → `myorg/myproject`
   - SSH: `git@gitlab.com:myorg/myproject.git` → `myorg/myproject`

### Filtering and Pagination

Most list operations support filtering and pagination:

```
# State filtering
List issues with state=opened for project myorg/myproject
List issues with state=closed for project myorg/myproject
List issues with state=all for project myorg/myproject

# Label filtering
List issues with labels "bug,critical" for project myorg/myproject

# Pagination
List issues with limit=50 for project myorg/myproject
List labels with limit=20 for project myorg/myproject
```

### Batch Operations

For bulk operations, use Claude Code's ability to chain commands:

```
1. List all open issues
2. For each issue with label "urgent", add a comment
3. Update status to high priority
```

---

## Error Handling

### Common Errors

**Project Not Found:**
```
Error: 404 Project Not Found
```
- Verify project path is correct
- Check you have access to the project
- Ensure token has appropriate scopes

**Issue Not Found:**
```
Error: Issue IID not found
```
- Use issue IID (internal ID), not global ID
- Verify issue exists in the specified project

**Permission Denied:**
```
Error: 403 Forbidden
```
- Check token has required scopes (`api`, `read_api`, `write_api`)
- Verify you have access to the project/group
- For epics, ensure you have Premium/Ultimate tier

**Label Validation Error:**
```
The following labels do not exist: 'typo-label'
```
- Use `list_labels` to see available labels
- Fix typos in label names
- Or set `GITLAB_VALIDATE_LABELS=false` to disable validation

**Rate Limiting:**
```
Error: 429 Too Many Requests
```
- Wait before retrying
- Reduce request frequency
- Contact GitLab admin if persistent

### Debugging Tips

1. **Enable debug logging**: Check Claude Code logs for detailed error messages
2. **Test with simple requests**: Start with basic operations before complex ones
3. **Verify environment variables**: Ensure `GITLAB_TOKEN` and `GITLAB_URI` are set correctly
4. **Check token scopes**: Verify token has all required permissions
5. **Use GitLab web interface**: Confirm operations work in the web UI first

For more help, see the [Troubleshooting Guide](TROUBLESHOOTING.md).

---

## Next Steps

- [Setup and configuration →](SETUP.md)
- [Docker deployment →](DOCKER.md)
- [Development guide →](DEVELOPMENT.md)
