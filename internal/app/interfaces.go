package app

import (
	"gitlab.com/gitlab-org/api/client-go"
)

// ProjectsService interface for GitLab Projects operations.
type ProjectsService interface {
	GetProject(pid any, opt *gitlab.GetProjectOptions) (*gitlab.Project, *gitlab.Response, error)
	EditProject(pid any, opt *gitlab.EditProjectOptions) (*gitlab.Project, *gitlab.Response, error)
}

// GroupsService interface for GitLab Groups operations.
type GroupsService interface {
	GetGroup(gid any, opt *gitlab.GetGroupOptions) (*gitlab.Group, *gitlab.Response, error)
}

// IssuesService interface for GitLab Issues operations.
type IssuesService interface {
	ListProjectIssues(pid any, opt *gitlab.ListProjectIssuesOptions) ([]*gitlab.Issue, *gitlab.Response, error)
	CreateIssue(pid any, opt *gitlab.CreateIssueOptions) (*gitlab.Issue, *gitlab.Response, error)
	UpdateIssue(pid any, issue int64, opt *gitlab.UpdateIssueOptions) (*gitlab.Issue, *gitlab.Response, error)
	GetIssue(pid any, issue int) (*gitlab.Issue, *gitlab.Response, error)
}

// LabelsService interface for GitLab Labels operations.
type LabelsService interface {
	ListLabels(pid any, opt *gitlab.ListLabelsOptions) ([]*gitlab.Label, *gitlab.Response, error)
}

// UsersService interface for GitLab Users operations.
type UsersService interface {
	CurrentUser() (*gitlab.User, *gitlab.Response, error)
	ListUsers(opt *gitlab.ListUsersOptions) ([]*gitlab.User, *gitlab.Response, error)
}

// NotesService interface for GitLab Notes operations.
type NotesService interface {
	CreateIssueNote(
		pid any,
		issue int64,
		opt *gitlab.CreateIssueNoteOptions,
	) (*gitlab.Note, *gitlab.Response, error)
}

// EpicsService interface for GitLab Epics operations.
type EpicsService interface {
	ListGroupEpics(gid any, opt *gitlab.ListGroupEpicsOptions) ([]*gitlab.Epic, *gitlab.Response, error)
	CreateEpic(gid any, opt *gitlab.CreateEpicOptions) (*gitlab.Epic, *gitlab.Response, error)
}

// EpicIssuesService interface for GitLab Epic Issues operations.
type EpicIssuesService interface {
	AssignEpicIssue(gid any, epic, issue int64) (*gitlab.EpicIssueAssignment, *gitlab.Response, error)
}

// GitLabClient interface that provides access to all GitLab services.
type GitLabClient interface {
	Projects() ProjectsService
	Issues() IssuesService
	Labels() LabelsService
	Users() UsersService
	Notes() NotesService
	Groups() GroupsService
	Epics() EpicsService
	EpicIssues() EpicIssuesService
}
