package app

import (
	"gitlab.com/gitlab-org/api/client-go"
)

// ProjectsService interface for GitLab Projects operations.
type ProjectsService interface {
	GetProject(pid interface{}, opt *gitlab.GetProjectOptions) (*gitlab.Project, *gitlab.Response, error)
	EditProject(pid interface{}, opt *gitlab.EditProjectOptions) (*gitlab.Project, *gitlab.Response, error)
}

// IssuesService interface for GitLab Issues operations.
type IssuesService interface {
	ListProjectIssues(pid interface{}, opt *gitlab.ListProjectIssuesOptions) ([]*gitlab.Issue, *gitlab.Response, error)
	CreateIssue(pid interface{}, opt *gitlab.CreateIssueOptions) (*gitlab.Issue, *gitlab.Response, error)
	UpdateIssue(pid interface{}, issue int64, opt *gitlab.UpdateIssueOptions) (*gitlab.Issue, *gitlab.Response, error)
}

// LabelsService interface for GitLab Labels operations.
type LabelsService interface {
	ListLabels(pid interface{}, opt *gitlab.ListLabelsOptions) ([]*gitlab.Label, *gitlab.Response, error)
}

// UsersService interface for GitLab Users operations.
type UsersService interface {
	CurrentUser() (*gitlab.User, *gitlab.Response, error)
	ListUsers(opt *gitlab.ListUsersOptions) ([]*gitlab.User, *gitlab.Response, error)
}

// NotesService interface for GitLab Notes operations.
type NotesService interface {
	CreateIssueNote(pid interface{}, issue int64, opt *gitlab.CreateIssueNoteOptions) (*gitlab.Note, *gitlab.Response, error)
}

// MergeRequestsService interface for GitLab MergeRequests operations.
type MergeRequestsService interface {
	CreateMergeRequest(
		pid interface{}, 
		opt *gitlab.CreateMergeRequestOptions,
	) (*gitlab.MergeRequest, *gitlab.Response, error)
}

// MilestonesService interface for GitLab Milestones operations.
type MilestonesService interface {
	ListMilestones(pid interface{}, opt *gitlab.ListMilestonesOptions) ([]*gitlab.Milestone, *gitlab.Response, error)
}

// GitLabClient interface that provides access to all GitLab services.
type GitLabClient interface {
	Projects() ProjectsService
	Issues() IssuesService
	Labels() LabelsService
	Users() UsersService
	Notes() NotesService
	MergeRequests() MergeRequestsService
	Milestones() MilestonesService
}