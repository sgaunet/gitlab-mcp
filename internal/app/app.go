package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sgaunet/gitlab-mcp/internal/logger"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Constants for default values.
const (
	defaultGitLabURI   = "https://gitlab.com/"
	defaultStateOpened = "opened"
	maxIssuesPerPage   = 100
	maxLabelsPerPage   = 100
)

// Error variables for static errors.
var (
	ErrGitLabTokenRequired   = errors.New("GITLAB_TOKEN environment variable is required")
	ErrCreateOptionsRequired = errors.New("create issue options are required")
	ErrIssueTitleRequired    = errors.New("issue title is required")
	ErrInvalidIssueIID       = errors.New("issue IID must be a positive integer")
	ErrUpdateOptionsRequired = errors.New("update issue options are required")
	ErrNoteBodyRequired      = errors.New("note body is required")
	ErrLabelValidationFailed = errors.New("label validation failed")
	ErrEpicsTierRequired     = errors.New("epics require GitLab Premium or Ultimate tier")
	ErrEpicTitleRequired     = errors.New("epic title is required")
	ErrInvalidDateFormat     = errors.New("date must be in YYYY-MM-DD format")
	ErrEpicIIDRequired       = errors.New("epic IID is required")
	ErrProjectPathRequired   = errors.New("project path is required")
	ErrGroupPathRequired     = errors.New("group path is required")
	ErrIssueNotFound         = errors.New("issue not found")
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

// ListEpicsOptions contains options for listing group epics.
type ListEpicsOptions struct {
	State string
	Limit int64
}

// CreateEpicOptions contains options for creating a group epic.
type CreateEpicOptions struct {
	Title        string
	Description  string
	Labels       []string
	StartDate    string
	DueDate      string
	Confidential bool
}

// AddIssueToEpicOptions contains options for attaching an issue to an epic.
type AddIssueToEpicOptions struct {
	GroupPath   string
	EpicIID     int
	ProjectPath string
	IssueIID    int64
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

// Epic represents a GitLab epic.
type Epic struct {
	ID          int64          `json:"id"`
	IID         int64          `json:"iid"`
	GroupID     int64          `json:"group_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	State       string         `json:"state"`
	WebURL      string         `json:"web_url"`
	Author      map[string]any `json:"author"`
	StartDate   string         `json:"start_date"`
	DueDate     string         `json:"due_date"`
	Labels      []string       `json:"labels"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

// EpicIssueAssignment represents an issue associated with an epic.
type EpicIssueAssignment struct {
	ID          int64          `json:"id"`
	IID         int64          `json:"iid"`
	EpicID      int64          `json:"epic_id"`
	EpicIID     int64          `json:"epic_iid"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	State       string         `json:"state"`
	WebURL      string         `json:"web_url"`
	Labels      []string       `json:"labels"`
	Author      map[string]any `json:"author"`
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


// convertGitLabEpic converts a GitLab epic to our Epic struct.
func convertGitLabEpic(epic *gitlab.Epic) Epic {
	// Convert author to the expected format
	var author map[string]any
	if epic.Author != nil {
		author = map[string]any{
			"id":       epic.Author.ID,
			"username": epic.Author.Username,
			"name":     epic.Author.Name,
		}
	}

	// Format dates (handle nil pointers)
	var startDate, dueDate, createdAt, updatedAt string
	if epic.StartDate != nil {
		startDate = epic.StartDate.String()
	}
	if epic.DueDate != nil {
		dueDate = epic.DueDate.String()
	}
	if epic.CreatedAt != nil {
		createdAt = epic.CreatedAt.Format("2006-01-02T15:04:05Z")
	}
	if epic.UpdatedAt != nil {
		updatedAt = epic.UpdatedAt.Format("2006-01-02T15:04:05Z")
	}

	return Epic{
		ID:          epic.ID,
		IID:         epic.IID,
		GroupID:     epic.GroupID,
		Title:       epic.Title,
		Description: epic.Description,
		State:       epic.State,
		WebURL:      epic.WebURL,
		Author:      author,
		StartDate:   startDate,
		DueDate:     dueDate,
		Labels:      epic.Labels,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// convertGitLabEpicIssueAssignment converts a GitLab EpicIssueAssignment to our EpicIssueAssignment struct.
func convertGitLabEpicIssueAssignment(epicIssue *gitlab.EpicIssueAssignment) EpicIssueAssignment {
	// The EpicIssueAssignment type in the GitLab client is an embedded Issue
	// with additional epic-related fields

	// Convert author to the expected format
	var author map[string]any
	if epicIssue.Issue.Author != nil {
		author = map[string]any{
			"id":       epicIssue.Issue.Author.ID,
			"username": epicIssue.Issue.Author.Username,
			"name":     epicIssue.Issue.Author.Name,
		}
	}

	return EpicIssueAssignment{
		ID:          epicIssue.Issue.ID,
		IID:         epicIssue.Issue.IID,
		EpicID:      epicIssue.Epic.ID,
		EpicIID:     epicIssue.Epic.IID,
		Title:       epicIssue.Issue.Title,
		Description: epicIssue.Issue.Description,
		State:       epicIssue.Issue.State,
		WebURL:      epicIssue.Issue.WebURL,
		Labels:      epicIssue.Issue.Labels,
		Author:      author,
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

	createNote := func(projectID int64, iid int64, body string) (*gitlab.Note, error) {
		createOpts := &gitlab.CreateIssueNoteOptions{Body: &body}
		note, _, err := a.client.Notes().CreateIssueNote(projectID, iid, createOpts)
		if err != nil {
			return nil, fmt.Errorf("gitlab API call failed: %w", err)
		}
		return note, nil
	}
	return a.addNoteCommon(projectPath, issueIID, opts.Body, "issue", createNote)
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

// convertGitLabNote converts a GitLab note to our Note struct.
func convertGitLabNote(note *gitlab.Note) *Note {
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

	return result
}

// ListGroupEpics retrieves epics for a given group path.
func (a *App) ListGroupEpics(groupPath string, opts *ListEpicsOptions) ([]Epic, error) {
	a.logger.Debug("Listing epics for group", "group_path", groupPath, "options", opts)

	// Resolve group path to group ID
	groupID, err := a.resolveGroupID(groupPath)
	if err != nil {
		return nil, err
	}

	// Set defaults for options
	opts = a.setDefaultEpicOptions(opts)

	// Create GitLab API options
	listOpts := &gitlab.ListGroupEpicsOptions{
		State:       &opts.State,
		ListOptions: gitlab.ListOptions{PerPage: opts.Limit, Page: 1},
	}

	// Call GitLab API
	epics, _, err := a.client.Epics().ListGroupEpics(groupID, listOpts)
	if err != nil {
		return nil, a.handleEpicsAPIError(err, groupID, "failed to list group epics")
	}

	a.logger.Debug("Retrieved epics", "count", len(epics), "group_id", groupID)

	// Convert GitLab epics to our Epic struct
	result := make([]Epic, 0, len(epics))
	for _, epic := range epics {
		result = append(result, convertGitLabEpic(epic))
	}

	a.logger.Info("Successfully retrieved group epics", "count", len(result), "group_id", groupID)
	return result, nil
}

// CreateGroupEpic creates a new epic in a GitLab group.
func (a *App) CreateGroupEpic(groupPath string, opts *CreateEpicOptions) (*Epic, error) {
	// Validate options
	if err := a.validateCreateEpicOptions(opts); err != nil {
		return nil, err
	}

	// Parse dates if provided (validate before API calls)
	startDate, dueDate, err := a.parseEpicDates(opts)
	if err != nil {
		return nil, err
	}

	a.logger.Debug("Creating epic for group", "group_path", groupPath, "title", opts.Title)

	// Resolve group path to group ID
	groupID, err := a.resolveGroupID(groupPath)
	if err != nil {
		return nil, err
	}

	// Build GitLab API options
	createOpts := a.buildCreateEpicOptions(opts, startDate, dueDate)

	// Call GitLab API
	epic, _, err := a.client.Epics().CreateEpic(groupID, createOpts)
	if err != nil {
		return nil, a.handleEpicsAPIError(err, groupID, "failed to create epic")
	}

	a.logger.Debug("Created epic", "id", epic.ID, "iid", epic.IID, "group_id", groupID)

	result := convertGitLabEpic(epic)
	a.logger.Info("Successfully created epic",
		"id", result.ID, "iid", result.IID, "group_id", groupID, "title", result.Title)

	return &result, nil
}

// AddIssueToEpic attaches an issue to an epic.
func (a *App) AddIssueToEpic(opts *AddIssueToEpicOptions) (*EpicIssueAssignment, error) {
	// Validate options
	if err := a.validateAddIssueToEpicOptions(opts); err != nil {
		return nil, err
	}

	a.logger.Debug("Adding issue to epic",
		"group_path", opts.GroupPath, "epic_iid", opts.EpicIID,
		"project_path", opts.ProjectPath, "issue_iid", opts.IssueIID)

	// Resolve group path to group ID
	groupID, err := a.resolveGroupID(opts.GroupPath)
	if err != nil {
		return nil, err
	}

	// Get project to ensure it exists
	project, _, err := a.client.Projects().GetProject(opts.ProjectPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", opts.ProjectPath, err)
	}

	// Get issue to obtain global issue ID
	issue, _, err := a.client.Issues().GetIssue(project.ID, int(opts.IssueIID))
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %d: %w", opts.IssueIID, err)
	}

	if issue.ID == 0 {
		return nil, ErrIssueNotFound
	}

	// Assign issue to epic
	epicIssue, _, err := a.client.EpicIssues().AssignEpicIssue(groupID, int64(opts.EpicIID), issue.ID)
	if err != nil {
		return nil, a.handleEpicsAPIError(err, groupID, "failed to add issue to epic")
	}

	a.logger.Debug("Added issue to epic", "epic_iid", opts.EpicIID, "issue_id", issue.ID)

	result := convertGitLabEpicIssueAssignment(epicIssue)
	a.logger.Info("Successfully added issue to epic",
		"issue_id", result.ID, "issue_iid", result.IID, "epic_iid", result.EpicIID)

	return &result, nil
}

// parseDate parses a date string in YYYY-MM-DD format to gitlab.ISOTime.
func (a *App) parseDate(dateStr string) (*gitlab.ISOTime, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidDateFormat, err)
	}
	isoTime := gitlab.ISOTime(t)
	return &isoTime, nil
}

// resolveGroupID resolves a group path to a group ID.
func (a *App) resolveGroupID(groupPath string) (int64, error) {
	group, _, err := a.client.Groups().GetGroup(groupPath, nil)
	if err != nil {
		a.logger.Error("Failed to get group", "error", err, "group_path", groupPath)

		// Check if it's a 403 Forbidden (Premium/Ultimate tier requirement)
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "Forbidden") {
			return 0, fmt.Errorf(
				"%w: group access may require Premium/Ultimate tier or epics feature is not enabled",
				ErrEpicsTierRequired,
			)
		}

		return 0, fmt.Errorf("failed to get group: %w", err)
	}
	return group.ID, nil
}

// setDefaultEpicOptions sets default values for epic listing options.
func (a *App) setDefaultEpicOptions(opts *ListEpicsOptions) *ListEpicsOptions {
	if opts == nil {
		return &ListEpicsOptions{
			State: "opened",
			Limit: maxIssuesPerPage,
		}
	}
	if opts.State == "" {
		opts.State = "opened"
	}
	if opts.Limit == 0 {
		opts.Limit = maxIssuesPerPage
	}
	if opts.Limit > maxIssuesPerPage {
		opts.Limit = maxIssuesPerPage
	}
	return opts
}

// handleEpicsAPIError handles errors from GitLab epics API calls.
func (a *App) handleEpicsAPIError(err error, groupID int64, context string) error {
	a.logger.Error(context, "error", err, "group_id", groupID)

	// Check if it's a 403 Forbidden (Premium/Ultimate tier requirement)
	if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "Forbidden") {
		return fmt.Errorf("%w: epics are only available in GitLab Premium or Ultimate tier", ErrEpicsTierRequired)
	}

	return fmt.Errorf("%s: %w", context, err)
}

// validateCreateEpicOptions validates epic creation options.
func (a *App) validateCreateEpicOptions(opts *CreateEpicOptions) error {
	if opts == nil {
		return ErrCreateOptionsRequired
	}
	if opts.Title == "" {
		return ErrEpicTitleRequired
	}
	return nil
}

// validateAddIssueToEpicOptions validates the options for adding an issue to an epic.
func (a *App) validateAddIssueToEpicOptions(opts *AddIssueToEpicOptions) error {
	if opts == nil {
		return ErrCreateOptionsRequired
	}
	if opts.GroupPath == "" {
		return ErrGroupPathRequired
	}
	if opts.EpicIID <= 0 {
		return ErrEpicIIDRequired
	}
	if opts.ProjectPath == "" {
		return ErrProjectPathRequired
	}
	if opts.IssueIID <= 0 {
		return ErrInvalidIssueIID
	}
	return nil
}

// parseEpicDates parses start and due dates from options.
func (a *App) parseEpicDates(opts *CreateEpicOptions) (*gitlab.ISOTime, *gitlab.ISOTime, error) {
	var startDate, dueDate *gitlab.ISOTime
	var err error

	if opts.StartDate != "" {
		startDate, err = a.parseDate(opts.StartDate)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid start_date: %w", err)
		}
	}

	if opts.DueDate != "" {
		dueDate, err = a.parseDate(opts.DueDate)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid due_date: %w", err)
		}
	}

	return startDate, dueDate, nil
}

// buildCreateEpicOptions builds GitLab API options from our options struct.
func (a *App) buildCreateEpicOptions(
	opts *CreateEpicOptions,
	startDate, dueDate *gitlab.ISOTime,
) *gitlab.CreateEpicOptions {
	createOpts := &gitlab.CreateEpicOptions{
		Title:       &opts.Title,
		Description: &opts.Description,
	}

	// Set dates with fixed flag
	if startDate != nil {
		createOpts.StartDateFixed = startDate
		fixed := true
		createOpts.StartDateIsFixed = &fixed
	}
	if dueDate != nil {
		createOpts.DueDateFixed = dueDate
		fixed := true
		createOpts.DueDateIsFixed = &fixed
	}

	// Set labels if provided
	if len(opts.Labels) > 0 {
		labels := gitlab.LabelOptions(opts.Labels)
		createOpts.Labels = &labels
	}

	// Set confidential if true
	if opts.Confidential {
		createOpts.Confidential = &opts.Confidential
	}

	return createOpts
}

// addNoteCommon handles common logic for adding notes.
func (a *App) addNoteCommon(
	projectPath string,
	iid int64,
	body string,
	noteType string,
	createNote func(projectID int64, iid int64, body string) (*gitlab.Note, error),
) (*Note, error) {
	a.logger.Debug("Adding note", "type", noteType, "project_path", projectPath, "iid", iid)

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	projectID := project.ID

	// Call GitLab API using the provided function
	note, err := createNote(projectID, iid, body)
	if err != nil {
		a.logger.Error("Failed to create note", "type", noteType, "error", err,
			"project_id", projectID, "iid", iid)
		return nil, fmt.Errorf("failed to create %s note: %w", noteType, err)
	}

	a.logger.Debug("Created note", "type", noteType, "id", note.ID,
		"project_id", projectID, "iid", iid)

	result := convertGitLabNote(note)

	a.logger.Info("Successfully added note", "type", noteType, "note_id", result.ID,
		"project_id", projectID, "iid", iid)
	return result, nil
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
