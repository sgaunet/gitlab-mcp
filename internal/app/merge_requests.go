package app

import (
	"errors"
	"fmt"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Error variables for merge request operations.
var (
	ErrMRTitleRequired      = errors.New("merge request title is required")
	ErrSourceBranchRequired = errors.New("source branch is required")
	ErrTargetBranchRequired = errors.New("target branch is required")
	ErrMROptionsRequired    = errors.New("merge request options are required")
	ErrInvalidMRIID         = errors.New("merge request IID must be positive")
)

const (
	maxMergeRequestsPerPage = 100
)

// buildListMergeRequestsOptions converts our options to GitLab API options.
func buildListMergeRequestsOptions(opts *ListMergeRequestsOptions) *gitlab.ListProjectMergeRequestsOptions {
	listOpts := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{PerPage: maxMergeRequestsPerPage, Page: 1},
	}

	if opts == nil {
		return listOpts
	}

	if opts.State != "" {
		listOpts.State = gitlab.Ptr(opts.State)
	}
	if len(opts.Labels) > 0 {
		labels := gitlab.LabelOptions(opts.Labels)
		listOpts.Labels = &labels
	}
	if opts.Author != "" {
		listOpts.AuthorUsername = gitlab.Ptr(opts.Author)
	}
	if opts.Search != "" {
		listOpts.Search = gitlab.Ptr(opts.Search)
	}
	if opts.Limit > 0 {
		listOpts.PerPage = min(opts.Limit, maxMergeRequestsPerPage)
	}

	return listOpts
}

// buildAcceptMergeRequestOptions converts our options to GitLab API options.
func buildAcceptMergeRequestOptions(opts *MergeMergeRequestOptions) *gitlab.AcceptMergeRequestOptions {
	acceptOpts := &gitlab.AcceptMergeRequestOptions{}

	if opts == nil {
		return acceptOpts
	}

	if opts.MergeCommitMessage != "" {
		acceptOpts.MergeCommitMessage = gitlab.Ptr(opts.MergeCommitMessage)
	}
	if opts.SquashCommitMessage != "" {
		acceptOpts.SquashCommitMessage = gitlab.Ptr(opts.SquashCommitMessage)
	}
	if opts.Squash {
		acceptOpts.Squash = gitlab.Ptr(opts.Squash)
	}
	if opts.ShouldRemoveSourceBranch {
		acceptOpts.ShouldRemoveSourceBranch = gitlab.Ptr(opts.ShouldRemoveSourceBranch)
	}
	if opts.MergeWhenPipelineSucceeds {
		acceptOpts.AutoMerge = gitlab.Ptr(opts.MergeWhenPipelineSucceeds)
	}

	return acceptOpts
}

// buildCreateMergeRequestOptions converts our options to GitLab API options.
func buildCreateMergeRequestOptions(opts *CreateMergeRequestOptions) *gitlab.CreateMergeRequestOptions {
	createOpts := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(opts.Title),
		SourceBranch: gitlab.Ptr(opts.SourceBranch),
		TargetBranch: gitlab.Ptr(opts.TargetBranch),
	}

	if opts.Description != "" {
		createOpts.Description = gitlab.Ptr(opts.Description)
	}
	if len(opts.Labels) > 0 {
		labels := gitlab.LabelOptions(opts.Labels)
		createOpts.Labels = &labels
	}
	if len(opts.AssigneeIDs) > 0 {
		createOpts.AssigneeIDs = &opts.AssigneeIDs
	}
	if len(opts.ReviewerIDs) > 0 {
		createOpts.ReviewerIDs = &opts.ReviewerIDs
	}

	return createOpts
}

// buildUpdateMergeRequestOptions converts our options to GitLab API options.
func buildUpdateMergeRequestOptions(opts *UpdateMergeRequestOptions) *gitlab.UpdateMergeRequestOptions {
	updateOpts := &gitlab.UpdateMergeRequestOptions{}

	if opts.Title != "" {
		updateOpts.Title = gitlab.Ptr(opts.Title)
	}
	if opts.Description != "" {
		updateOpts.Description = gitlab.Ptr(opts.Description)
	}
	if opts.State != "" {
		updateOpts.StateEvent = gitlab.Ptr(opts.State)
	}
	if opts.Labels != nil {
		labels := gitlab.LabelOptions(opts.Labels)
		updateOpts.Labels = &labels
	}
	if opts.AssigneeIDs != nil {
		updateOpts.AssigneeIDs = &opts.AssigneeIDs
	}
	if opts.ReviewerIDs != nil {
		updateOpts.ReviewerIDs = &opts.ReviewerIDs
	}

	return updateOpts
}

// ListProjectMergeRequests lists merge requests for a GitLab project.
func (a *App) ListProjectMergeRequests(
	projectPath string,
	opts *ListMergeRequestsOptions,
) ([]MergeRequest, error) {
	a.logger.Debug("Listing merge requests", "project_path", projectPath, "options", opts)

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project %s: %w", projectPath, err)
	}
	projectID := project.ID

	// Build GitLab API options
	listOpts := buildListMergeRequestsOptions(opts)

	// Call GitLab API
	mrs, _, err := a.client.MergeRequests().ListProjectMergeRequests(projectID, listOpts)
	if err != nil {
		a.logger.Error("Failed to list merge requests", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("failed to list merge requests for project %s: %w", projectPath, err)
	}

	// Convert to our MergeRequest type
	result := make([]MergeRequest, 0, len(mrs))
	for _, mr := range mrs {
		result = append(result, convertGitLabMergeRequest(mr))
	}

	a.logger.Info("Successfully listed merge requests", "count", len(result), "project_path", projectPath)
	return result, nil
}

// CreateProjectMergeRequest creates a new merge request.
func (a *App) CreateProjectMergeRequest(
	projectPath string,
	opts *CreateMergeRequestOptions,
) (*MergeRequest, error) {
	// Validate options
	if opts == nil {
		return nil, ErrMROptionsRequired
	}

	a.logger.Debug("Creating merge request", "project_path", projectPath, "title", opts.Title)
	if opts.Title == "" {
		return nil, ErrMRTitleRequired
	}
	if opts.SourceBranch == "" {
		return nil, ErrSourceBranchRequired
	}
	if opts.TargetBranch == "" {
		return nil, ErrTargetBranchRequired
	}

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project %s: %w", projectPath, err)
	}
	projectID := project.ID

	// Build GitLab API options
	createOpts := buildCreateMergeRequestOptions(opts)

	// Call GitLab API
	mr, _, err := a.client.MergeRequests().CreateMergeRequest(projectID, createOpts)
	if err != nil {
		a.logger.Error("Failed to create merge request", "error", err, "project_id", projectID)
		return nil, fmt.Errorf("failed to create merge request in project %s: %w", projectPath, err)
	}

	result := convertGitLabMergeRequest(mr)
	a.logger.Info("Successfully created merge request", "mr_iid", result.IID, "project_path", projectPath)
	return &result, nil
}

// GetMergeRequest gets details of a specific merge request.
func (a *App) GetMergeRequest(projectPath string, mrIID int64) (*MergeRequest, error) {
	a.logger.Debug("Getting merge request", "project_path", projectPath, "mr_iid", mrIID)

	if mrIID <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidMRIID, mrIID)
	}

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project %s: %w", projectPath, err)
	}
	projectID := project.ID

	// Call GitLab API
	mr, _, err := a.client.MergeRequests().GetMergeRequest(projectID, int(mrIID), nil)
	if err != nil {
		a.logger.Error("Failed to get merge request", "error", err, "mr_iid", mrIID)
		return nil, fmt.Errorf("failed to get merge request %d in project %s: %w", mrIID, projectPath, err)
	}

	result := convertGitLabMergeRequest(mr)
	a.logger.Info("Successfully retrieved merge request", "mr_iid", mrIID, "project_path", projectPath)
	return &result, nil
}

// UpdateMergeRequest updates an existing merge request.
func (a *App) UpdateMergeRequest(
	projectPath string,
	mrIID int64,
	opts *UpdateMergeRequestOptions,
) (*MergeRequest, error) {
	a.logger.Debug("Updating merge request", "project_path", projectPath, "mr_iid", mrIID)

	if mrIID <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidMRIID, mrIID)
	}
	if opts == nil {
		return nil, ErrMROptionsRequired
	}

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project %s: %w", projectPath, err)
	}
	projectID := project.ID

	// Build GitLab API options
	updateOpts := buildUpdateMergeRequestOptions(opts)

	// Call GitLab API
	mr, _, err := a.client.MergeRequests().UpdateMergeRequest(projectID, int(mrIID), updateOpts)
	if err != nil {
		a.logger.Error("Failed to update merge request", "error", err, "mr_iid", mrIID)
		return nil, fmt.Errorf("failed to update merge request %d in project %s: %w", mrIID, projectPath, err)
	}

	result := convertGitLabMergeRequest(mr)
	a.logger.Info("Successfully updated merge request", "mr_iid", mrIID, "project_path", projectPath)
	return &result, nil
}

// MergeMergeRequest merges an approved merge request.
func (a *App) MergeMergeRequest(
	projectPath string,
	mrIID int64,
	opts *MergeMergeRequestOptions,
) (*MergeRequest, error) {
	a.logger.Debug("Merging merge request", "project_path", projectPath, "mr_iid", mrIID)

	if mrIID <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidMRIID, mrIID)
	}

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return nil, fmt.Errorf("failed to get project %s: %w", projectPath, err)
	}
	projectID := project.ID

	// Build GitLab API options
	acceptOpts := buildAcceptMergeRequestOptions(opts)

	// Call GitLab API
	mr, _, err := a.client.MergeRequests().AcceptMergeRequest(projectID, int(mrIID), acceptOpts)
	if err != nil {
		a.logger.Error("Failed to merge request", "error", err, "mr_iid", mrIID)
		return nil, fmt.Errorf("failed to merge request %d in project %s: %w", mrIID, projectPath, err)
	}

	result := convertGitLabMergeRequest(mr)
	a.logger.Info("Successfully merged merge request", "mr_iid", mrIID, "project_path", projectPath)
	return &result, nil
}

// GetMergeRequestDiff gets the diff for a merge request.
func (a *App) GetMergeRequestDiff(projectPath string, mrIID int64) (string, error) {
	a.logger.Debug("Getting merge request diff", "project_path", projectPath, "mr_iid", mrIID)

	if mrIID <= 0 {
		return "", fmt.Errorf("%w: got %d", ErrInvalidMRIID, mrIID)
	}

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return "", fmt.Errorf("failed to get project %s: %w", projectPath, err)
	}
	projectID := project.ID

	// Call GitLab API to get actual file-level diffs
	diffs, _, err := a.client.MergeRequests().ListMergeRequestDiffs(projectID, int(mrIID), nil)
	if err != nil {
		a.logger.Error("Failed to get merge request diff", "error", err, "mr_iid", mrIID)
		return "", fmt.Errorf("failed to get diff for merge request %d in project %s: %w", mrIID, projectPath, err)
	}

	if len(diffs) == 0 {
		return "No diffs available for this merge request", nil
	}

	// Format diffs into unified diff output
	var result strings.Builder
	for i, diff := range diffs {
		if i > 0 {
			result.WriteString("\n")
		}
		switch {
		case diff.NewFile:
			fmt.Fprintf(&result, "--- /dev/null\n+++ b/%s\n", diff.NewPath)
		case diff.DeletedFile:
			fmt.Fprintf(&result, "--- a/%s\n+++ /dev/null\n", diff.OldPath)
		default:
			fmt.Fprintf(&result, "--- a/%s\n+++ b/%s\n", diff.OldPath, diff.NewPath)
		}
		result.WriteString(diff.Diff)
	}

	a.logger.Info("Successfully retrieved merge request diff",
		"mr_iid", mrIID, "project_path", projectPath, "files_changed", len(diffs))
	return result.String(), nil
}

// ApproveMergeRequest approves a merge request.
func (a *App) ApproveMergeRequest(
	projectPath string,
	mrIID int64,
	opts *ApproveMergeRequestOptions,
) (string, error) {
	a.logger.Debug("Approving merge request", "project_path", projectPath, "mr_iid", mrIID)

	if mrIID <= 0 {
		return "", fmt.Errorf("%w: got %d", ErrInvalidMRIID, mrIID)
	}

	// Get project by path
	project, _, err := a.client.Projects().GetProject(projectPath, nil)
	if err != nil {
		a.logger.Error("Failed to get project", "error", err, "project_path", projectPath)
		return "", fmt.Errorf("failed to get project %s: %w", projectPath, err)
	}
	projectID := project.ID

	// Build GitLab API options
	approveOpts := &gitlab.ApproveMergeRequestOptions{}
	if opts != nil && opts.SHA != "" {
		approveOpts.SHA = gitlab.Ptr(opts.SHA)
	}

	// Call GitLab API
	approvals, _, err := a.client.MergeRequestApprovals().ApproveMergeRequest(projectID, mrIID, approveOpts)
	if err != nil {
		a.logger.Error("Failed to approve merge request", "error", err, "mr_iid", mrIID)
		return "", fmt.Errorf("failed to approve merge request %d in project %s: %w", mrIID, projectPath, err)
	}

	result := fmt.Sprintf("Merge request %d approved successfully. Approvals: %d/%d",
		mrIID, len(approvals.ApprovedBy), approvals.ApprovalsRequired)

	a.logger.Info("Successfully approved merge request", "mr_iid", mrIID, "project_path", projectPath)
	return result, nil
}

// AddMergeRequestNote adds a comment/note to a merge request.
func (a *App) AddMergeRequestNote(
	projectPath string,
	mrIID int64,
	opts *AddMergeRequestNoteOptions,
) (*Note, error) {
	a.logger.Debug("Adding merge request note", "project_path", projectPath, "mr_iid", mrIID)

	if opts == nil || opts.Body == "" {
		return nil, ErrNoteBodyRequired
	}
	if mrIID <= 0 {
		return nil, fmt.Errorf("%w: got %d", ErrInvalidMRIID, mrIID)
	}

	// Use common note creation helper
	return a.addNoteCommon(projectPath, mrIID, opts.Body, "merge request",
		func(projectID int64, iid int64, body string) (*gitlab.Note, error) {
			createOpts := &gitlab.CreateMergeRequestNoteOptions{
				Body: gitlab.Ptr(body),
			}
			note, _, err := a.client.Notes().CreateMergeRequestNote(projectID, iid, createOpts)
			if err != nil {
				return nil, fmt.Errorf("failed to create merge request note: %w", err)
			}
			return note, nil
		})
}

// convertGitLabMergeRequest converts a GitLab merge request to our MergeRequest struct.
func convertGitLabMergeRequest(mr *gitlab.MergeRequest) MergeRequest {
	result := MergeRequest{
		ID:             mr.ID,
		IID:            mr.IID,
		Title:          mr.Title,
		Description:    mr.Description,
		State:          mr.State,
		SourceBranch:   mr.SourceBranch,
		TargetBranch:   mr.TargetBranch,
		Labels:         mr.Labels,
		WebURL:         mr.WebURL,
		MergeStatus:    mr.DetailedMergeStatus,
		Draft:          mr.Draft,
		WorkInProgress: mr.Draft,  // Draft is the new field
	}

	// Convert author
	if mr.Author != nil {
		result.Author = map[string]any{
			"id":       mr.Author.ID,
			"username": mr.Author.Username,
			"name":     mr.Author.Name,
		}
	}

	// Convert assignees
	if len(mr.Assignees) > 0 {
		result.Assignees = make([]map[string]any, 0, len(mr.Assignees))
		for _, assignee := range mr.Assignees {
			result.Assignees = append(result.Assignees, map[string]any{
				"id":       assignee.ID,
				"username": assignee.Username,
				"name":     assignee.Name,
			})
		}
	}

	// Convert reviewers
	if len(mr.Reviewers) > 0 {
		result.Reviewers = make([]map[string]any, 0, len(mr.Reviewers))
		for _, reviewer := range mr.Reviewers {
			result.Reviewers = append(result.Reviewers, map[string]any{
				"id":       reviewer.ID,
				"username": reviewer.Username,
				"name":     reviewer.Name,
			})
		}
	}

	// Convert timestamps
	if mr.CreatedAt != nil {
		result.CreatedAt = mr.CreatedAt.Format("2006-01-02T15:04:05Z")
	}
	if mr.UpdatedAt != nil {
		result.UpdatedAt = mr.UpdatedAt.Format("2006-01-02T15:04:05Z")
	}

	return result
}
