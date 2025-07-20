package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/sgaunet/gitlab-mcp/internal/logger"
	"gitlab.com/gitlab-org/api/client-go"
)

// Constants for default values.
const (
	defaultGitLabURI    = "https://gitlab.com/"
	defaultStateOpened  = "opened"
	maxIssuesPerPage    = 100
	maxLabelsPerPage    = 100
)

// Error variables for static errors.
var (
	ErrGitLabTokenRequired   = errors.New("GITLAB_TOKEN environment variable is required")
	ErrCreateOptionsRequired = errors.New("create issue options are required")
	ErrIssueTitleRequired    = errors.New("issue title is required")
	ErrInvalidIssueIID       = errors.New("issue IID must be a positive integer")
	ErrUpdateOptionsRequired = errors.New("update issue options are required")
)

type App struct {
	GitLabToken string
	GitLabURI   string
	client      GitLabClient
	logger      *slog.Logger
}

func New() (*App, error) {
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		return nil, ErrGitLabTokenRequired
	}

	uri := os.Getenv("GITLAB_URI")
	if uri == "" {
		uri = defaultGitLabURI
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
		client:      NewGitLabClient(client),
		logger:      logger.NoLogger(),
	}, nil
}

// NewWithClient creates a new App instance with an injected GitLabClient (for testing).
func NewWithClient(token, uri string, client GitLabClient) *App {
	return &App{
		GitLabToken: token,
		GitLabURI:   uri,
		client:      client,
		logger:      logger.NoLogger(),
	}
}

func (a *App) GetAPIURL() string {
	return a.GitLabURI + "/api/v4"
}

func (a *App) SetLogger(l *slog.Logger) {
	a.logger = l
}

func (a *App) ValidateConnection() error {
	_, _, err := a.client.Users().CurrentUser()
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	
	return nil
}


// ListIssuesOptions contains options for listing project issues.
type ListIssuesOptions struct {
	State  string
	Labels string
	Limit  int
}

// CreateIssueOptions contains options for creating a project issue.
type CreateIssueOptions struct {
	Title       string
	Description string
	Labels      []string
	Assignees   []int
}

// UpdateIssueOptions contains options for updating a project issue.
type UpdateIssueOptions struct {
	Title       string
	Description string
	State       string
	Labels      []string
	Assignees   []int
}

// ListLabelsOptions contains options for listing project labels.
type ListLabelsOptions struct {
	WithCounts            bool
	IncludeAncestorGroups bool
	Search                string
	Limit                 int
}

// Issue represents a GitLab issue.
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

// Label represents a GitLab label.
type Label struct {
	ID                     int    `json:"id"`
	Name                   string `json:"name"`
	Color                  string `json:"color"`
	TextColor              string `json:"text_color"`
	Description            string `json:"description"`
	OpenIssuesCount        int    `json:"open_issues_count"`
	ClosedIssuesCount      int    `json:"closed_issues_count"`
	OpenMergeRequestsCount int    `json:"open_merge_requests_count"`
	Subscribed             bool   `json:"subscribed"`
	Priority               int    `json:"priority"`
	IsProjectLabel         bool   `json:"is_project_label"`
}

// parseLabels splits comma-separated labels and trims spaces.
func parseLabels(labels string) []string {
	parts := strings.Split(labels, ",")
	result := make([]string, 0, len(parts))
	for _, label := range parts {
		trimmed := strings.TrimSpace(label)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// convertGitLabIssue converts a GitLab issue to our Issue struct.
func convertGitLabIssue(issue *gitlab.Issue) Issue {
	// Convert assignees to the expected format
	assignees := make([]map[string]interface{}, 0, len(issue.Assignees))
	for _, assignee := range issue.Assignees {
		assignees = append(assignees, map[string]interface{}{
			"id":       assignee.ID,
			"username": assignee.Username,
			"name":     assignee.Name,
		})
	}

	return Issue{
		ID:          issue.ID,
		IID:         issue.IID,
		Title:       issue.Title,
		Description: issue.Description,
		State:       issue.State,
		Labels:      issue.Labels,
		Assignees:   assignees,
		CreatedAt:   issue.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   issue.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// normalizeListIssuesOptions sets default values for list issues options.
func normalizeListIssuesOptions(opts *ListIssuesOptions) *ListIssuesOptions {
	if opts == nil {
		opts = &ListIssuesOptions{}
	}
	if opts.State == "" {
		opts.State = defaultStateOpened
	}
	if opts.Limit == 0 {
		opts.Limit = maxIssuesPerPage
	}
	if opts.Limit > maxIssuesPerPage {
		opts.Limit = maxIssuesPerPage
	}
	return opts
}

// ListProjectIssues retrieves issues for a given project path.
func (a *App) ListProjectIssues(projectPath string, opts *ListIssuesOptions) ([]Issue, error) {
	a.logger.Debug("Listing issues for project", "project_path", projectPath, "options", opts)
	
	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	projectID := project.ID

	// Normalize options
	opts = normalizeListIssuesOptions(opts)

	// Create GitLab API options
	listOpts := &gitlab.ListProjectIssuesOptions{
		State:       &opts.State,
		ListOptions: gitlab.ListOptions{PerPage: opts.Limit, Page: 1},
	}

	// Add labels filter if provided
	if opts.Labels != "" {
		// Split comma-separated labels and trim spaces
		labelList := parseLabels(opts.Labels)
		if len(labelList) > 0 {
			labels := gitlab.LabelOptions(labelList)
			listOpts.Labels = &labels
		}
	}

	// Call GitLab API
	issues, _, err := a.client.Issues().ListProjectIssues(projectID, listOpts)
	if err != nil {
		a.logger.Error("Failed to list project issues", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("failed to list project issues: %w", err)
	}

	a.logger.Debug("Retrieved issues", "count", len(issues), "project_id", projectID)

	// Convert GitLab issues to our Issue struct
	result := make([]Issue, 0, len(issues))
	for _, issue := range issues {
		result = append(result, convertGitLabIssue(issue))
	}

	a.logger.Info("Successfully retrieved project issues", "count", len(result), "project_id", projectID)
	return result, nil
}

// CreateProjectIssue creates a new issue for a given project path.
func (a *App) CreateProjectIssue(projectPath string, opts *CreateIssueOptions) (*Issue, error) {
	// Validate required options
	if opts == nil {
		return nil, ErrCreateOptionsRequired
	}
	if opts.Title == "" {
		return nil, ErrIssueTitleRequired
	}
	
	a.logger.Debug("Creating issue for project", "project_path", projectPath, "title", opts.Title)
	
	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	projectID := project.ID

	// Create GitLab API options
	createOpts := &gitlab.CreateIssueOptions{
		Title:       &opts.Title,
		Description: &opts.Description,
	}

	// Add labels if provided
	if len(opts.Labels) > 0 {
		labels := gitlab.LabelOptions(opts.Labels)
		createOpts.Labels = &labels
	}

	// Add assignees if provided
	if len(opts.Assignees) > 0 {
		createOpts.AssigneeIDs = &opts.Assignees
	}

	// Call GitLab API
	issue, _, err := a.client.Issues().CreateIssue(projectID, createOpts)
	if err != nil {
		a.logger.Error("Failed to create issue", "error", err, "project_id", projectID, "title", opts.Title)
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	a.logger.Debug("Created issue", "id", issue.ID, "iid", issue.IID, "project_id", projectID)

	result := convertGitLabIssue(issue)
	a.logger.Info("Successfully created issue", 
		"id", result.ID, 
		"iid", result.IID, 
		"project_id", projectID, 
		"title", result.Title)
	return &result, nil
}

// ListProjectLabels retrieves labels for a given project path.
func (a *App) ListProjectLabels(projectPath string, opts *ListLabelsOptions) ([]Label, error) {
	a.logger.Debug("Listing labels for project", "project_path", projectPath, "options", opts)
	
	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	projectID := project.ID

	// Set default options if not provided
	if opts == nil {
		opts = &ListLabelsOptions{
			WithCounts:            false,
			IncludeAncestorGroups: false,
			Limit:                 maxLabelsPerPage,
		}
	}

	// Set defaults for individual options
	if opts.Limit == 0 {
		opts.Limit = maxLabelsPerPage
	}
	if opts.Limit > maxLabelsPerPage {
		opts.Limit = maxLabelsPerPage // Cap at max labels per page
	}

	// Create GitLab API options
	listOpts := &gitlab.ListLabelsOptions{
		WithCounts:            &opts.WithCounts,
		IncludeAncestorGroups: &opts.IncludeAncestorGroups,
		ListOptions:           gitlab.ListOptions{PerPage: opts.Limit, Page: 1},
	}

	// Add search filter if provided
	if opts.Search != "" {
		listOpts.Search = &opts.Search
	}

	// Call GitLab API
	labels, _, err := a.client.Labels().ListLabels(projectID, listOpts)
	if err != nil {
		a.logger.Error("Failed to list project labels", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("failed to list project labels: %w", err)
	}

	a.logger.Debug("Retrieved labels", "count", len(labels), "project_id", projectID)

	// Convert GitLab labels to our Label struct
	result := make([]Label, 0, len(labels))
	for _, label := range labels {
		result = append(result, Label{
			ID:                     label.ID,
			Name:                   label.Name,
			Color:                  label.Color,
			TextColor:              label.TextColor,
			Description:            label.Description,
			OpenIssuesCount:        label.OpenIssuesCount,
			ClosedIssuesCount:      label.ClosedIssuesCount,
			OpenMergeRequestsCount: label.OpenMergeRequestsCount,
			Subscribed:             label.Subscribed,
			Priority:               label.Priority,
			IsProjectLabel:         label.IsProjectLabel,
		})
	}

	a.logger.Info("Successfully retrieved project labels", "count", len(result), "project_id", projectID)
	return result, nil
}

// UpdateProjectIssue updates an existing issue for a given project path.
func (a *App) UpdateProjectIssue(projectPath string, issueIID int, opts *UpdateIssueOptions) (*Issue, error) {
	// Validate required parameters
	if issueIID <= 0 {
		return nil, ErrInvalidIssueIID
	}
	if opts == nil {
		return nil, ErrUpdateOptionsRequired
	}
	
	a.logger.Debug("Updating issue for project", "project_path", projectPath, "issue_iid", issueIID)
	
	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	projectID := project.ID

	// Create GitLab API options - only set fields that are provided
	updateOpts := &gitlab.UpdateIssueOptions{}

	if opts.Title != "" {
		updateOpts.Title = &opts.Title
	}
	
	if opts.Description != "" {
		updateOpts.Description = &opts.Description
	}
	
	if opts.State != "" {
		updateOpts.StateEvent = &opts.State
	}

	// Add labels if provided
	if len(opts.Labels) > 0 {
		labels := gitlab.LabelOptions(opts.Labels)
		updateOpts.Labels = &labels
	}

	// Add assignees if provided
	if len(opts.Assignees) > 0 {
		updateOpts.AssigneeIDs = &opts.Assignees
	}

	// Call GitLab API
	issue, _, err := a.client.Issues().UpdateIssue(projectID, issueIID, updateOpts)
	if err != nil {
		a.logger.Error("Failed to update issue", "error", err, "project_id", projectID, "issue_iid", issueIID)
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	a.logger.Debug("Updated issue", "id", issue.ID, "iid", issue.IID, "project_id", projectID)

	result := convertGitLabIssue(issue)
	a.logger.Info("Successfully updated issue", 
		"id", result.ID, 
		"iid", result.IID, 
		"project_id", projectID, 
		"title", result.Title)
	return &result, nil
}
