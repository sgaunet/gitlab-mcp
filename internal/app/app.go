package app

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/sgaunet/gitlab-mcp/internal/logger"
	"gitlab.com/gitlab-org/api/client-go"
)

type App struct {
	GitLabToken string
	GitLabURI   string
	client      *gitlab.Client
	logger      *slog.Logger
}

func New() (*App, error) {
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITLAB_TOKEN environment variable is required")
	}

	uri := os.Getenv("GITLAB_URI")
	if uri == "" {
		uri = "https://gitlab.com/"
	}

	if _, err := url.Parse(uri); err != nil {
		return nil, fmt.Errorf("invalid GitLab URI: %w", err)
	}

	var client *gitlab.Client
	var err error
	if uri == "https://gitlab.com/" {
		client, err = gitlab.NewClient(token)
	} else {
		client, err = gitlab.NewClient(token, gitlab.WithBaseURL(uri))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &App{
		GitLabToken: token,
		GitLabURI:   uri,
		client:      client,
		logger:      logger.NoLogger(),
	}, nil
}

func (a *App) GetAPIURL() string {
	return fmt.Sprintf("%s/api/v4", a.GitLabURI)
}

func (a *App) SetLogger(l *slog.Logger) {
	a.logger = l
}

func (a *App) ValidateConnection() error {
	_, _, err := a.client.Users.CurrentUser()
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	
	return nil
}

func (a *App) GetProjectID(remoteURL string) (int, error) {
	a.logger.Debug("Getting project ID for remote URL", "url", remoteURL)
	
	projectName, err := a.extractProjectName(remoteURL)
	if err != nil {
		a.logger.Error("Failed to extract project name", "error", err, "url", remoteURL)
		return 0, fmt.Errorf("failed to extract project name: %w", err)
	}
	
	a.logger.Debug("Extracted project name", "name", projectName)
	
	searchOpts := &gitlab.SearchOptions{}
	foundProjects, _, err := a.client.Search.Projects(projectName, searchOpts)
	if err != nil {
		a.logger.Error("Failed to search projects", "error", err, "project_name", projectName)
		return 0, fmt.Errorf("failed to search projects: %w", err)
	}
	
	a.logger.Debug("Found projects", "count", len(foundProjects))
	
	for _, p := range foundProjects {
		a.logger.Debug("Checking project", "id", p.ID, "ssh_url", p.SSHURLToRepo, "http_url", p.HTTPURLToRepo)
		if p.SSHURLToRepo == remoteURL || p.HTTPURLToRepo == remoteURL {
			a.logger.Info("Found matching project", "id", p.ID, "name", p.Name)
			return p.ID, nil
		}
	}
	
	a.logger.Warn("Project not found", "url", remoteURL)
	return 0, fmt.Errorf("project not found for remote URL: %s", remoteURL)
}

func (a *App) extractProjectName(remoteURL string) (string, error) {
	remoteURL = strings.TrimSpace(remoteURL)
	
	var projectPath string
	
	if strings.HasPrefix(remoteURL, "git@") {
		parts := strings.Split(remoteURL, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid SSH URL format")
		}
		projectPath = strings.TrimSuffix(parts[1], ".git")
	} else if strings.HasPrefix(remoteURL, "http://") || strings.HasPrefix(remoteURL, "https://") {
		u, err := url.Parse(remoteURL)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
		projectPath = strings.TrimPrefix(u.Path, "/")
		projectPath = strings.TrimSuffix(projectPath, ".git")
	} else {
		return "", fmt.Errorf("unsupported remote URL format")
	}
	
	parts := strings.Split(projectPath, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid project path format")
	}
	
	return parts[len(parts)-1], nil
}

// ListIssuesOptions contains options for listing project issues
type ListIssuesOptions struct {
	State  string
	Labels string
	Limit  int
}

// Issue represents a GitLab issue
type Issue struct {
	ID          int                    `json:"id"`
	IID         int                    `json:"iid"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	State       string                 `json:"state"`
	Labels      []string               `json:"labels"`
	Assignees   []map[string]interface{} `json:"assignees"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// ListProjectIssues retrieves issues for a given project ID
func (a *App) ListProjectIssues(projectID int, opts *ListIssuesOptions) ([]Issue, error) {
	a.logger.Debug("Listing issues for project", "project_id", projectID, "options", opts)

	// Set default options if not provided
	if opts == nil {
		opts = &ListIssuesOptions{
			State: "opened",
			Limit: 100,
		}
	}

	// Set defaults for individual options
	if opts.State == "" {
		opts.State = "opened"
	}
	if opts.Limit == 0 {
		opts.Limit = 100
	}
	if opts.Limit > 100 {
		opts.Limit = 100 // Cap at 100 issues
	}

	// Create GitLab API options
	listOpts := &gitlab.ListProjectIssuesOptions{
		State:       &opts.State,
		ListOptions: gitlab.ListOptions{PerPage: opts.Limit, Page: 1},
	}

	// TODO: Add labels filter support once we understand the correct type
	// For now, labels filtering is not implemented

	// Call GitLab API
	issues, _, err := a.client.Issues.ListProjectIssues(projectID, listOpts)
	if err != nil {
		a.logger.Error("Failed to list project issues", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("failed to list project issues: %w", err)
	}

	a.logger.Debug("Retrieved issues", "count", len(issues), "project_id", projectID)

	// Convert GitLab issues to our Issue struct
	result := make([]Issue, 0, len(issues))
	for _, issue := range issues {
		// Convert assignees to the expected format
		assignees := make([]map[string]interface{}, 0, len(issue.Assignees))
		for _, assignee := range issue.Assignees {
			assignees = append(assignees, map[string]interface{}{
				"id":       assignee.ID,
				"username": assignee.Username,
				"name":     assignee.Name,
			})
		}

		result = append(result, Issue{
			ID:          issue.ID,
			IID:         issue.IID,
			Title:       issue.Title,
			Description: issue.Description,
			State:       issue.State,
			Labels:      issue.Labels,
			Assignees:   assignees,
			CreatedAt:   issue.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   issue.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	a.logger.Info("Successfully retrieved project issues", "count", len(result), "project_id", projectID)
	return result, nil
}
