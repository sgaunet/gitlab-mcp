# Contributing Guide

[< Back to README](../README.md)

Thank you for your interest in contributing to gitlab-mcp! This guide will help you get started.

## Table of Contents

- [How to Contribute](#how-to-contribute)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Commit Message Conventions](#commit-message-conventions)
- [Issue Reporting](#issue-reporting)
- [Pull Request Process](#pull-request-process)
- [Review Process](#review-process)
- [License Agreement](#license-agreement)

---

## How to Contribute

There are many ways to contribute to gitlab-mcp:

### Code Contributions
- Fix bugs
- Add new features
- Improve performance
- Add new GitLab tools
- Enhance documentation

### Non-Code Contributions
- Report bugs
- Suggest features
- Improve documentation
- Write tutorials
- Answer questions in issues
- Share your experience

### Areas That Need Help
- Testing on different platforms
- Docker deployment improvements
- Performance optimization
- Documentation examples
- Error message improvements

---

## Development Workflow

### 1. Fork the Repository

Click the "Fork" button on the [GitHub repository](https://github.com/sgaunet/gitlab-mcp).

### 2. Clone Your Fork

```bash
git clone https://github.com/YOUR_USERNAME/gitlab-mcp.git
cd gitlab-mcp
```

### 3. Add Upstream Remote

```bash
git remote add upstream https://github.com/sgaunet/gitlab-mcp.git
git remote -v
```

### 4. Create a Feature Branch

```bash
# Update your fork
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/my-new-feature

# Or for bug fixes
git checkout -b fix/bug-description
```

### 5. Make Your Changes

```bash
# Edit files
vim internal/app/app.go

# Run tests
task test

# Run linter
task lint

# Build to verify
task build
```

### 6. Commit Your Changes

Follow the [commit message conventions](#commit-message-conventions):

```bash
git add .
git commit -m "feat: add new GitLab tool for merge requests"
```

### 7. Push to Your Fork

```bash
git push origin feature/my-new-feature
```

### 8. Create a Pull Request

- Go to your fork on GitHub
- Click "Compare & pull request"
- Fill out the PR template
- Submit the pull request

---

## Coding Standards

### Go Style Guide

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments):

- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Write clear, idiomatic Go code
- Add comments for exported functions and types

### Code Organization

**File Structure:**
```go
// Package declaration
package app

// Imports (standard library, then third-party, then internal)
import (
    "context"
    "errors"
    "fmt"

    "github.com/xanzy/go-gitlab"

    "github.com/sgaunet/gitlab-mcp/internal/logger"
)

// Constants
const (
    DefaultLimit = 100
)

// Error variables
var (
    ErrInvalidInput = errors.New("invalid input")
)

// Types

// Functions
```

### Error Handling

**Use static error variables:**
```go
var (
    ErrProjectPathRequired = errors.New("project_path is required")
    ErrInvalidStateValue   = errors.New("state must be 'opened' or 'closed'")
)
```

**Wrap errors with context:**
```go
if err != nil {
    return nil, fmt.Errorf("failed to get project: %w", err)
}
```

**Don't panic in library code:**
```go
// ❌ Bad
if err != nil {
    panic(err)
}

// ✅ Good
if err != nil {
    return fmt.Errorf("failed to process: %w", err)
}
```

### Logging

**Use structured logging:**
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

**Log levels:**
- `Debug`: Detailed diagnostic information
- `Info`: General informational messages
- `Warn`: Warning messages for recoverable issues
- `Error`: Error messages for failures

### Testing

**Write tests for all public methods:**
```go
func TestListProjectIssues(t *testing.T) {
    // Setup
    mockClient := &MockGitLabClient{}
    app := NewWithClient(mockClient, cfg, logger)

    // Expectations
    // ...

    // Execute
    result, err := app.ListProjectIssues(...)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

**Use table-driven tests:**
```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{
    {"valid input", "test", "test", false},
    {"empty input", "", "", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := Func(tt.input)
        if (err != nil) != tt.wantErr {
            t.Errorf("wanted error: %v, got: %v", tt.wantErr, err)
        }
        if got != tt.want {
            t.Errorf("wanted: %v, got: %v", tt.want, got)
        }
    })
}
```

### Documentation

**Comment exported items:**
```go
// ListProjectIssues retrieves issues for a GitLab project by project path.
// It supports filtering by state, labels, and pagination.
//
// Parameters:
//   - projectPath: GitLab project path (e.g., "namespace/project")
//   - opts: Optional filtering and pagination options
//
// Returns:
//   - Slice of gitlab.Issue objects
//   - Error if the operation fails
func (a *App) ListProjectIssues(projectPath string, opts *ListIssuesOptions) ([]*gitlab.Issue, error) {
    // ...
}
```

**Update documentation files:**
- Update README.md if adding user-facing features
- Update docs/TOOLS.md when adding new tools
- Update docs/DEVELOPMENT.md for architectural changes
- Add examples to CLAUDE.md for AI assistance

---

## Commit Message Conventions

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, no logic change)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks (dependencies, build, etc.)
- **perf**: Performance improvements
- **ci**: CI/CD changes

### Examples

**Feature:**
```
feat(issues): add support for filtering by milestone

Add milestone parameter to list_issues tool for filtering
issues by milestone ID or title.

Closes #42
```

**Bug Fix:**
```
fix(labels): correct label validation error message

The error message was unclear when labels didn't exist.
Now includes list of available labels.

Fixes #38
```

**Documentation:**
```
docs(docker): add troubleshooting section

Add common Docker issues and solutions to DOCKER.md
```

**Refactor:**
```
refactor(app): extract project validation logic

Extract project path validation into a separate function
for reuse across multiple tools.
```

### Rules

- Use imperative mood ("add" not "added")
- Don't capitalize first letter of subject
- No period at the end of subject
- Limit subject line to 50 characters
- Separate subject from body with blank line
- Wrap body at 72 characters
- Reference issues in footer

---

## Issue Reporting

### Bug Reports

When reporting bugs, include:

**Environment:**
```
- OS: macOS 14.1
- Go version: 1.21.5
- Installation method: Homebrew
- GitLab version: GitLab.com (or self-hosted version)
```

**Steps to Reproduce:**
```
1. Install gitlab-mcp via Homebrew
2. Set GITLAB_TOKEN environment variable
3. Run command: `List all issues for project myorg/myproject`
4. Observe error: ...
```

**Expected Behavior:**
```
Should list all open issues
```

**Actual Behavior:**
```
Returns error: "404 Project Not Found"
```

**Logs/Output:**
```
Include relevant logs or error messages
```

**Additional Context:**
```
Any other information that might help
```

### Feature Requests

When requesting features, include:

**Problem Statement:**
```
What problem does this solve?
```

**Proposed Solution:**
```
Describe your ideal solution
```

**Alternatives Considered:**
```
Other approaches you've thought about
```

**Additional Context:**
```
Examples, mockups, or related issues
```

### Security Issues

**DO NOT** create public issues for security vulnerabilities.

Instead:
1. Email security@example.com (if available)
2. Or create a private security advisory on GitHub
3. Include full details and proof of concept
4. Allow time for patch before public disclosure

---

## Pull Request Process

### Before Submitting

- [ ] Code follows style guidelines
- [ ] Tests pass: `task test`
- [ ] Linter passes: `task lint`
- [ ] Documentation updated
- [ ] Commit messages follow conventions
- [ ] Branch is up to date with main

### PR Template

Fill out the pull request template:

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How to test these changes

## Checklist
- [ ] Tests pass
- [ ] Linter passes
- [ ] Documentation updated
- [ ] Commit messages follow conventions
```

### PR Guidelines

**Good PR:**
- Single focused change
- Clear description
- Tests included
- Documentation updated
- Small, reviewable size

**PR to Avoid:**
- Multiple unrelated changes
- No description
- No tests
- Very large diffs
- Breaking changes without discussion

---

## Review Process

### What to Expect

1. **Automated Checks** (~5 minutes)
   - Linter runs
   - Tests execute
   - Coverage calculated

2. **Maintainer Review** (1-3 days)
   - Code quality review
   - Architecture feedback
   - Suggestions for improvement

3. **Revisions** (as needed)
   - Address feedback
   - Update based on comments
   - Re-request review

4. **Approval & Merge** (~1 day)
   - Approved by maintainer
   - Merged to main
   - Included in next release

### Review Criteria

**Code Quality:**
- Follows Go best practices
- Clear and maintainable
- Well-tested
- Properly documented

**Functionality:**
- Solves the stated problem
- No breaking changes (without discussion)
- Edge cases handled
- Error handling appropriate

**Testing:**
- Unit tests included
- Tests cover new code
- Tests pass consistently
- Coverage maintained or improved

**Documentation:**
- User-facing docs updated
- Code comments added
- Examples provided
- CHANGELOG updated (if applicable)

---

## License Agreement

By contributing to gitlab-mcp, you agree that your contributions will be licensed under the MIT License.

You confirm that:
- You own the copyright or have permission to contribute
- Your contribution doesn't violate any third-party rights
- You grant the project permission to use your contribution

---

## Recognition

Contributors are recognized in:
- GitHub contributors list
- Release notes (for significant contributions)
- Project documentation (for major features)

---

## Questions?

- **Documentation**: Check [docs/](.)
- **Issues**: Search [existing issues](https://github.com/sgaunet/gitlab-mcp/issues)
- **Discussions**: Ask in GitHub Discussions (if available)

Thank you for contributing to gitlab-mcp! 🎉

---

## Next Steps

- [Setup development environment →](DEVELOPMENT.md)
- [Learn about the architecture →](DEVELOPMENT.md#project-architecture)
- [Browse open issues →](https://github.com/sgaunet/gitlab-mcp/issues)
