package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/sgaunet/gitlab-mcp/internal/logger"
	"gitlab.com/gitlab-org/api/client-go"
)

// Constants for default values.
const (
	defaultGitLabURI     = "https://gitlab.com/"
	defaultStateOpened   = "opened"
	maxIssuesPerPage     = 100
	maxLabelsPerPage     = 100
	maxMilestonesPerPage = 100
)

// Error variables for static errors.
var (
	ErrGitLabTokenRequired            = errors.New("GITLAB_TOKEN environment variable is required")
	ErrCreateOptionsRequired          = errors.New("create issue options are required")
	ErrIssueTitleRequired             = errors.New("issue title is required")
	ErrInvalidIssueIID                = errors.New("issue IID must be a positive integer")
	ErrUpdateOptionsRequired          = errors.New("update issue options are required")
	ErrNoteBodyRequired               = errors.New("note body is required")
	ErrCreateMROptionsRequired        = errors.New("create merge request options are required")
	ErrMRTitleRequired                = errors.New("merge request title is required")
	ErrMRSourceBranchRequired         = errors.New("merge request source branch is required")
	ErrMRTargetBranchRequired         = errors.New("merge request target branch is required")
	ErrInvalidUserIdentifierType      = errors.New("invalid user identifier type")
	ErrUserNotFound                   = errors.New("user not found")
	ErrInvalidMilestoneIdentifierType = errors.New("invalid milestone identifier type")
	ErrMilestoneNotFound              = errors.New("milestone not found")
	ErrLabelValidationFailed          = errors.New("label validation failed")
)

type App struct {
	GitLabToken    string
	GitLabURI      string
	ValidateLabels bool
	client         GitLabClient
	logger         *slog.Logger
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

	// Parse validate labels setting (default: true)
	validateLabels := true
	if validateStr := os.Getenv("GITLAB_VALIDATE_LABELS"); validateStr != "" {
		if parsed, err := strconv.ParseBool(validateStr); err == nil {
			validateLabels = parsed
		}
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
		GitLabToken:    token,
		GitLabURI:      uri,
		ValidateLabels: validateLabels,
		client:         NewGitLabClient(client),
		logger:         logger.NoLogger(),
	}, nil
}

// NewWithClient creates a new App instance with an injected GitLabClient (for testing).
func NewWithClient(token, uri string, client GitLabClient) *App {
	return &App{
		GitLabToken:    token,
		GitLabURI:      uri,
		ValidateLabels: true, // default for tests
		client:         client,
		logger:         logger.NoLogger(),
	}
}

// NewWithClientAndValidation creates a new App instance with an injected GitLabClient and
// validation setting (for testing).
func NewWithClientAndValidation(token, uri string, client GitLabClient, validateLabels bool) *App {
	return &App{
		GitLabToken:    token,
		GitLabURI:      uri,
		ValidateLabels: validateLabels,
		client:         client,
		logger:         logger.NoLogger(),
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
	Limit  int64
}

// CreateIssueOptions contains options for creating a project issue.
type CreateIssueOptions struct {
	Title       string
	Description string
	Labels      []string
	Assignees   []int64
}

// UpdateIssueOptions contains options for updating a project issue.
type UpdateIssueOptions struct {
	Title       string
	Description string
	State       string
	Labels      []string
	Assignees   []int64
}

// ListLabelsOptions contains options for listing project labels.
type ListLabelsOptions struct {
	WithCounts            bool
	IncludeAncestorGroups bool
	Search                string
	Limit                 int64
}

// AddIssueNoteOptions contains options for adding a note to an issue.
type AddIssueNoteOptions struct {
	Body string
}

// CreateMergeRequestOptions contains options for creating a merge request.
type CreateMergeRequestOptions struct {
	SourceBranch       string
	TargetBranch       string
	Title              string
	Description        string
	Assignees          []any // Can be usernames (string) or IDs (int)
	Reviewers          []any // Can be usernames (string) or IDs (int)
	Labels             []string
	Milestone          any // Can be title (string) or ID (int)
	RemoveSourceBranch bool
	Draft              bool
}

// Issue represents a GitLab issue.
type Issue struct {
	ID          int64            `json:"id"`
	IID         int64            `json:"iid"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	State       string           `json:"state"`
	Labels      []string         `json:"labels"`
	Assignees   []map[string]any `json:"assignees"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

// Label represents a GitLab label.
type Label struct {
	ID                     int64  `json:"id"`
	Name                   string `json:"name"`
	Color                  string `json:"color"`
	TextColor              string `json:"text_color"`
	Description            string `json:"description"`
	OpenIssuesCount        int64  `json:"open_issues_count"`
	ClosedIssuesCount      int64  `json:"closed_issues_count"`
	OpenMergeRequestsCount int64  `json:"open_merge_requests_count"`
	Subscribed             bool   `json:"subscribed"`
	Priority               int64  `json:"priority"`
	IsProjectLabel         bool   `json:"is_project_label"`
}

// Note represents a GitLab note/comment.
type Note struct {
	ID        int64          `json:"id"`
	Body      string         `json:"body"`
	Author    map[string]any `json:"author"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
	System    bool           `json:"system"`
	Noteable  map[string]any `json:"noteable"`
}

// MergeRequest represents a GitLab merge request.
type MergeRequest struct {
	ID           int64            `json:"id"`
	IID          int64            `json:"iid"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	State        string           `json:"state"`
	SourceBranch string           `json:"source_branch"`
	TargetBranch string           `json:"target_branch"`
	Author       map[string]any   `json:"author"`
	Assignees    []map[string]any `json:"assignees"`
	Reviewers    []map[string]any `json:"reviewers"`
	Labels       []string         `json:"labels"`
	Milestone    map[string]any   `json:"milestone"`
	WebURL       string           `json:"web_url"`
	Draft        bool             `json:"draft"`
	CreatedAt    string           `json:"created_at"`
	UpdatedAt    string           `json:"updated_at"`
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
	assignees := make([]map[string]any, 0, len(issue.Assignees))
	for _, assignee := range issue.Assignees {
		assignees = append(assignees, map[string]any{
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

// convertGitLabMergeRequest converts a GitLab merge request to our MergeRequest struct.
func convertGitLabMergeRequest(mr *gitlab.MergeRequest) MergeRequest {
	// Convert assignees to the expected format
	assignees := make([]map[string]any, 0, len(mr.Assignees))
	for _, assignee := range mr.Assignees {
		assignees = append(assignees, map[string]any{
			"id":       assignee.ID,
			"username": assignee.Username,
			"name":     assignee.Name,
		})
	}

	// Convert reviewers to the expected format
	reviewers := make([]map[string]any, 0, len(mr.Reviewers))
	for _, reviewer := range mr.Reviewers {
		reviewers = append(reviewers, map[string]any{
			"id":       reviewer.ID,
			"username": reviewer.Username,
			"name":     reviewer.Name,
		})
	}

	// Convert author to the expected format
	var author map[string]any
	if mr.Author != nil {
		author = map[string]any{
			"id":       mr.Author.ID,
			"username": mr.Author.Username,
			"name":     mr.Author.Name,
		}
	}

	// Convert milestone to the expected format
	var milestone map[string]any
	if mr.Milestone != nil {
		milestone = map[string]any{
			"id":    mr.Milestone.ID,
			"title": mr.Milestone.Title,
		}
	}

	return MergeRequest{
		ID:           mr.ID,
		IID:          mr.IID,
		Title:        mr.Title,
		Description:  mr.Description,
		State:        mr.State,
		SourceBranch: mr.SourceBranch,
		TargetBranch: mr.TargetBranch,
		Author:       author,
		Assignees:    assignees,
		Reviewers:    reviewers,
		Labels:       mr.Labels,
		Milestone:    milestone,
		WebURL:       mr.WebURL,
		Draft:        mr.Draft,
		CreatedAt:    mr.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    mr.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

	// Validate labels if validation is enabled
	if a.ValidateLabels && len(opts.Labels) > 0 {
		if err := a.validateLabels(projectID, projectPath, opts.Labels); err != nil {
			return nil, err
		}
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
func (a *App) UpdateProjectIssue(projectPath string, issueIID int64, opts *UpdateIssueOptions) (*Issue, error) {
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

// AddIssueNote adds a note/comment to an existing issue.
func (a *App) AddIssueNote(projectPath string, issueIID int64, opts *AddIssueNoteOptions) (*Note, error) {
	// Validate required parameters
	if issueIID <= 0 {
		return nil, ErrInvalidIssueIID
	}
	if opts == nil || opts.Body == "" {
		return nil, ErrNoteBodyRequired
	}

	a.logger.Debug("Adding note to issue", "project_path", projectPath, "issue_iid", issueIID)

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	projectID := project.ID

	// Create GitLab API options
	createOpts := &gitlab.CreateIssueNoteOptions{
		Body: &opts.Body,
	}

	// Call GitLab API
	note, _, err := a.client.Notes().CreateIssueNote(projectID, issueIID, createOpts)
	if err != nil {
		a.logger.Error("Failed to create issue note", "error", err, "project_id", projectID, "issue_iid", issueIID)
		return nil, fmt.Errorf("failed to create issue note: %w", err)
	}

	a.logger.Debug("Created issue note", "id", note.ID, "project_id", projectID, "issue_iid", issueIID)

	// Convert GitLab note to our Note struct
	result := &Note{
		ID:        note.ID,
		Body:      note.Body,
		System:    note.System,
		CreatedAt: note.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: note.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Convert author information
	if note.Author.ID != 0 {
		result.Author = map[string]any{
			"id":       note.Author.ID,
			"username": note.Author.Username,
			"name":     note.Author.Name,
		}
	}

	// Convert noteable information
	if note.NoteableID != 0 {
		result.Noteable = map[string]any{
			"id":   note.NoteableID,
			"iid":  note.NoteableIID,
			"type": note.NoteableType,
		}
	}

	a.logger.Info("Successfully added note to issue",
		"note_id", result.ID,
		"project_id", projectID,
		"issue_iid", issueIID)
	return result, nil
}

// CreateProjectMergeRequest creates a new merge request for a given project path.
func (a *App) CreateProjectMergeRequest(projectPath string, opts *CreateMergeRequestOptions) (*MergeRequest, error) {
	if err := a.validateMergeRequestOptions(opts); err != nil {
		return nil, err
	}

	a.logger.Debug("Creating merge request for project",
		"project_path", projectPath,
		"title", opts.Title,
		"source_branch", opts.SourceBranch,
		"target_branch", opts.TargetBranch)

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	projectID := project.ID

	// Create GitLab API options
	createOpts, err := a.buildMergeRequestOptions(projectID, opts)
	if err != nil {
		return nil, err
	}

	// Call GitLab API
	mr, _, err := a.client.MergeRequests().CreateMergeRequest(projectID, createOpts)
	if err != nil {
		a.logger.Error("Failed to create merge request",
			"error", err,
			"project_id", projectID,
			"title", opts.Title)
		return nil, fmt.Errorf("failed to create merge request: %w", err)
	}

	a.logger.Debug("Created merge request",
		"id", mr.ID,
		"iid", mr.IID,
		"project_id", projectID)

	result := convertGitLabMergeRequest(mr)
	a.logger.Info("Successfully created merge request",
		"id", result.ID,
		"iid", result.IID,
		"project_id", projectID,
		"title", result.Title)
	return &result, nil
}

// ProjectInfo represents basic project information.
type ProjectInfo struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Description string   `json:"description"`
	Topics      []string `json:"topics"`
}

// GetProjectDescription retrieves the description of a GitLab project.
func (a *App) GetProjectDescription(projectPath string) (*ProjectInfo, error) {
	a.logger.Debug("Getting project description", "project_path", projectPath)

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	result := &ProjectInfo{
		ID:          project.ID,
		Name:        project.Name,
		Path:        project.Path,
		Description: project.Description,
	}

	a.logger.Info("Successfully retrieved project description",
		"project_id", project.ID,
		"project_path", projectPath)
	return result, nil
}

// UpdateProjectDescription updates the description of a GitLab project.
func (a *App) UpdateProjectDescription(projectPath string, description string) (*ProjectInfo, error) {
	a.logger.Debug("Updating project description", "project_path", projectPath)

	// Get project by path first to get the ID
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	projectID := project.ID

	// Create update options
	updateOpts := &gitlab.EditProjectOptions{
		Description: &description,
	}

	// Update the project
	updatedProject, _, err := a.client.Projects().EditProject(projectID, updateOpts)
	if err != nil {
		a.logger.Error("Failed to update project description", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("failed to update project description: %w", err)
	}

	result := &ProjectInfo{
		ID:          updatedProject.ID,
		Name:        updatedProject.Name,
		Path:        updatedProject.Path,
		Description: updatedProject.Description,
		Topics:      updatedProject.Topics,
	}

	a.logger.Info("Successfully updated project description",
		"project_id", projectID,
		"project_path", projectPath)
	return result, nil
}

// GetProjectTopics retrieves the topics of a GitLab project.
func (a *App) GetProjectTopics(projectPath string) (*ProjectInfo, error) {
	a.logger.Debug("Getting project topics", "project_path", projectPath)

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	result := &ProjectInfo{
		ID:     project.ID,
		Name:   project.Name,
		Path:   project.Path,
		Topics: project.Topics,
	}

	a.logger.Info("Successfully retrieved project topics",
		"project_id", project.ID,
		"project_path", projectPath,
		"topics_count", len(project.Topics))
	return result, nil
}

// UpdateProjectTopics updates the topics of a GitLab project.
func (a *App) UpdateProjectTopics(projectPath string, topics []string) (*ProjectInfo, error) {
	a.logger.Debug("Updating project topics", "project_path", projectPath, "topics", topics)

	// Get project by path first to get the ID
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	projectID := project.ID

	// Create update options
	updateOpts := &gitlab.EditProjectOptions{
		Topics: &topics,
	}

	// Update the project
	updatedProject, _, err := a.client.Projects().EditProject(projectID, updateOpts)
	if err != nil {
		a.logger.Error("Failed to update project topics", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("failed to update project topics: %w", err)
	}

	result := &ProjectInfo{
		ID:          updatedProject.ID,
		Name:        updatedProject.Name,
		Path:        updatedProject.Path,
		Description: updatedProject.Description,
		Topics:      updatedProject.Topics,
	}

	a.logger.Info("Successfully updated project topics",
		"project_id", projectID,
		"project_path", projectPath,
		"topics_count", len(updatedProject.Topics))
	return result, nil
}

// validateMergeRequestOptions validates the required merge request options.
func (a *App) validateMergeRequestOptions(opts *CreateMergeRequestOptions) error {
	if opts == nil {
		return ErrCreateMROptionsRequired
	}
	if opts.Title == "" {
		return ErrMRTitleRequired
	}
	if opts.SourceBranch == "" {
		return ErrMRSourceBranchRequired
	}
	if opts.TargetBranch == "" {
		return ErrMRTargetBranchRequired
	}
	return nil
}

// buildMergeRequestOptions builds the GitLab API options for creating a merge request.
func (a *App) buildMergeRequestOptions(projectID int64, opts *CreateMergeRequestOptions) (
	*gitlab.CreateMergeRequestOptions, error,
) {
	createOpts := &gitlab.CreateMergeRequestOptions{
		Title:        &opts.Title,
		SourceBranch: &opts.SourceBranch,
		TargetBranch: &opts.TargetBranch,
	}

	// Add optional description
	if opts.Description != "" {
		createOpts.Description = &opts.Description
	}

	// Resolve assignees (usernames to IDs)
	if len(opts.Assignees) > 0 {
		assigneeIDs, err := a.resolveUserIdentifiers(opts.Assignees)
		if err != nil {
			a.logger.Error("Failed to resolve assignees", "error", err)
			return nil, fmt.Errorf("failed to resolve assignees: %w", err)
		}
		createOpts.AssigneeIDs = &assigneeIDs
	}

	// Resolve reviewers (usernames to IDs)
	if len(opts.Reviewers) > 0 {
		reviewerIDs, err := a.resolveUserIdentifiers(opts.Reviewers)
		if err != nil {
			a.logger.Error("Failed to resolve reviewers", "error", err)
			return nil, fmt.Errorf("failed to resolve reviewers: %w", err)
		}
		createOpts.ReviewerIDs = &reviewerIDs
	}

	// Add labels if provided
	if len(opts.Labels) > 0 {
		labels := gitlab.LabelOptions(opts.Labels)
		createOpts.Labels = &labels
	}

	// Resolve milestone (title to ID)
	if opts.Milestone != nil {
		milestoneID, err := a.resolveMilestoneIdentifier(projectID, opts.Milestone)
		if err != nil {
			a.logger.Error("Failed to resolve milestone", "error", err)
			return nil, fmt.Errorf("failed to resolve milestone: %w", err)
		}
		if milestoneID > 0 {
			createOpts.MilestoneID = &milestoneID
		}
	}

	// Set remove source branch option (default to true in issue spec)
	createOpts.RemoveSourceBranch = &opts.RemoveSourceBranch

	// Note: Draft is handled by GitLab automatically based on the title prefix "Draft:" or "WIP:"
	// The Draft field in our struct is for output only

	return createOpts, nil
}

// resolveUserIdentifiers converts username strings or IDs to user IDs.
func (a *App) resolveUserIdentifiers(identifiers []any) ([]int64, error) {
	if len(identifiers) == 0 {
		return nil, nil
	}

	userIDs := make([]int64, 0, len(identifiers))

	for _, identifier := range identifiers {
		switch v := identifier.(type) {
		case float64:
			// It's already an ID
			userIDs = append(userIDs, int64(v))
		case int:
			// It's already an ID
			userIDs = append(userIDs, int64(v))
		case string:
			// It's a username, need to resolve
			userID, err := a.findUserByUsername(v)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve user '%s': %w", v, err)
			}
			userIDs = append(userIDs, userID)
		default:
			return nil, fmt.Errorf("%w: %T", ErrInvalidUserIdentifierType, identifier)
		}
	}

	return userIDs, nil
}

// findUserByUsername searches for a user by username and returns their ID.
func (a *App) findUserByUsername(username string) (int64, error) {
	a.logger.Debug("Searching for user by username", "username", username)

	// Search for the user
	listOpts := &gitlab.ListUsersOptions{
		Username:    &username,
		ListOptions: gitlab.ListOptions{PerPage: 1, Page: 1},
	}

	users, _, err := a.client.Users().ListUsers(listOpts)
	if err != nil {
		a.logger.Error("Failed to search for user", "error", err, "username", username)
		return 0, fmt.Errorf("failed to search for user: %w", err)
	}

	if len(users) == 0 {
		a.logger.Error("User not found", "username", username)
		return 0, fmt.Errorf("%w: %s", ErrUserNotFound, username)
	}

	a.logger.Debug("Found user", "username", username, "id", users[0].ID)
	return users[0].ID, nil
}

// resolveMilestoneIdentifier converts milestone title or ID to milestone ID.
func (a *App) resolveMilestoneIdentifier(projectID int64, identifier any) (int64, error) {
	switch v := identifier.(type) {
	case float64:
		// It's already an ID
		return int64(v), nil
	case int:
		// It's already an ID
		return int64(v), nil
	case string:
		// It's a title, need to resolve
		return a.findMilestoneByTitle(projectID, v)
	default:
		return 0, fmt.Errorf("%w: %T", ErrInvalidMilestoneIdentifierType, identifier)
	}
}

// findMilestoneByTitle searches for a milestone by title and returns its ID.
func (a *App) findMilestoneByTitle(projectID int64, title string) (int64, error) {
	a.logger.Debug("Searching for milestone by title", "project_id", projectID, "title", title)

	// Search for active milestones
	state := "active"
	listOpts := &gitlab.ListMilestonesOptions{
		State:       &state,
		ListOptions: gitlab.ListOptions{PerPage: maxMilestonesPerPage, Page: 1},
	}

	milestones, _, err := a.client.Milestones().ListMilestones(projectID, listOpts)
	if err != nil {
		a.logger.Error("Failed to list milestones", "error", err, "project_id", projectID)
		return 0, fmt.Errorf("failed to list milestones: %w", err)
	}

	// Look for exact match
	for _, milestone := range milestones {
		if milestone.Title == title {
			a.logger.Debug("Found milestone", "title", title, "id", milestone.ID)
			return milestone.ID, nil
		}
	}

	a.logger.Error("Milestone not found", "title", title)
	return 0, fmt.Errorf("%w: %s", ErrMilestoneNotFound, title)
}

// validateLabels checks if the requested labels exist in the project.
func (a *App) validateLabels(projectID int64, projectPath string, requestedLabels []string) error {
	if len(requestedLabels) == 0 {
		return nil // No labels to validate
	}

	a.logger.Debug("Validating labels", "project_id", projectID, "requested_labels", requestedLabels)

	// Get existing labels from the project
	existingLabels, err := a.ListProjectLabels(projectPath, &ListLabelsOptions{
		Limit: maxLabelsPerPage,
	})
	if err != nil {
		a.logger.Error("Failed to retrieve existing labels for validation", "error", err, "project_id", projectID)
		return fmt.Errorf("failed to validate labels: %w", err)
	}

	// Create a map of existing label names (case-insensitive)
	existingLabelMap := make(map[string]bool)
	existingLabelNames := make([]string, 0, len(existingLabels))
	for _, label := range existingLabels {
		existingLabelMap[strings.ToLower(label.Name)] = true
		existingLabelNames = append(existingLabelNames, label.Name)
	}

	// Check which requested labels don't exist
	var missingLabels []string
	for _, requestedLabel := range requestedLabels {
		if !existingLabelMap[strings.ToLower(requestedLabel)] {
			missingLabels = append(missingLabels, requestedLabel)
		}
	}

	if len(missingLabels) > 0 {
		a.logger.Warn("Labels not found", "missing_labels", missingLabels, "project_path", projectPath)

		// Format error message with missing labels and available labels
		var errorMsg strings.Builder
		fmt.Fprintf(&errorMsg, "The following labels do not exist in project '%s':\n", projectPath)
		for _, label := range missingLabels {
			fmt.Fprintf(&errorMsg, "- '%s'\n", label)
		}

		if len(existingLabelNames) > 0 {
			errorMsg.WriteString("\nAvailable labels in this project:\n- ")
			errorMsg.WriteString(strings.Join(existingLabelNames, ", "))
		} else {
			errorMsg.WriteString("\nThis project has no labels defined.")
		}

		errorMsg.WriteString("\n\nTo disable label validation, set GITLAB_VALIDATE_LABELS=false")

		return fmt.Errorf("%w: %s", ErrLabelValidationFailed, errorMsg.String())
	}

	a.logger.Debug("All requested labels are valid", "project_id", projectID)
	return nil
}
