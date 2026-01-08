package app

import (
	"fmt"

	"gitlab.com/gitlab-org/api/client-go"
)

// GitLabClientWrapper wraps the real GitLab client to implement our interfaces.
type GitLabClientWrapper struct {
	client *gitlab.Client
}

// NewGitLabClient creates a new GitLab client wrapper.
func NewGitLabClient(client *gitlab.Client) GitLabClient {
	return &GitLabClientWrapper{
		client: client,
	}
}

// Projects returns the Projects service.
func (g *GitLabClientWrapper) Projects() ProjectsService {
	return &ProjectsServiceWrapper{service: g.client.Projects}
}

// Issues returns the Issues service.
func (g *GitLabClientWrapper) Issues() IssuesService {
	return &IssuesServiceWrapper{service: g.client.Issues}
}

// Labels returns the Labels service.
func (g *GitLabClientWrapper) Labels() LabelsService {
	return &LabelsServiceWrapper{service: g.client.Labels}
}

// Users returns the Users service.
func (g *GitLabClientWrapper) Users() UsersService {
	return &UsersServiceWrapper{service: g.client.Users}
}

// Notes returns the Notes service.
func (g *GitLabClientWrapper) Notes() NotesService {
	return &NotesServiceWrapper{service: g.client.Notes}
}

// MergeRequests returns the MergeRequests service.
func (g *GitLabClientWrapper) MergeRequests() MergeRequestsService {
	return &MergeRequestsServiceWrapper{service: g.client.MergeRequests}
}

// Milestones returns the Milestones service.
func (g *GitLabClientWrapper) Milestones() MilestonesService {
	return &MilestonesServiceWrapper{service: g.client.Milestones}
}

// Groups returns the Groups service.
func (g *GitLabClientWrapper) Groups() GroupsService {
	return &GroupsServiceWrapper{service: g.client.Groups}
}

// Epics returns the Epics service.
func (g *GitLabClientWrapper) Epics() EpicsService {
	return &EpicsServiceWrapper{service: g.client.Epics}
}

// ProjectsServiceWrapper wraps the real Projects service.
type ProjectsServiceWrapper struct {
	service gitlab.ProjectsServiceInterface
}

func (p *ProjectsServiceWrapper) GetProject(
	pid any,
	opt *gitlab.GetProjectOptions,
) (*gitlab.Project, *gitlab.Response, error) {
	project, resp, err := p.service.GetProject(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return project, resp, nil
}

func (p *ProjectsServiceWrapper) EditProject(
	pid any,
	opt *gitlab.EditProjectOptions,
) (*gitlab.Project, *gitlab.Response, error) {
	project, resp, err := p.service.EditProject(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return project, resp, nil
}

// IssuesServiceWrapper wraps the real Issues service.
type IssuesServiceWrapper struct {
	service gitlab.IssuesServiceInterface
}

func (i *IssuesServiceWrapper) ListProjectIssues(
	pid any,
	opt *gitlab.ListProjectIssuesOptions,
) ([]*gitlab.Issue, *gitlab.Response, error) {
	issues, resp, err := i.service.ListProjectIssues(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return issues, resp, nil
}

func (i *IssuesServiceWrapper) CreateIssue(
	pid any,
	opt *gitlab.CreateIssueOptions,
) (*gitlab.Issue, *gitlab.Response, error) {
	issue, resp, err := i.service.CreateIssue(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return issue, resp, nil
}

func (i *IssuesServiceWrapper) UpdateIssue(
	pid any,
	issue int64,
	opt *gitlab.UpdateIssueOptions,
) (*gitlab.Issue, *gitlab.Response, error) {
	updatedIssue, resp, err := i.service.UpdateIssue(pid, issue, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return updatedIssue, resp, nil
}

// LabelsServiceWrapper wraps the real Labels service.
type LabelsServiceWrapper struct {
	service gitlab.LabelsServiceInterface
}

func (l *LabelsServiceWrapper) ListLabels(
	pid any,
	opt *gitlab.ListLabelsOptions,
) ([]*gitlab.Label, *gitlab.Response, error) {
	labels, resp, err := l.service.ListLabels(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return labels, resp, nil
}

// UsersServiceWrapper wraps the real Users service.
type UsersServiceWrapper struct {
	service gitlab.UsersServiceInterface
}

func (u *UsersServiceWrapper) CurrentUser() (*gitlab.User, *gitlab.Response, error) {
	user, resp, err := u.service.CurrentUser()
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return user, resp, nil
}

func (u *UsersServiceWrapper) ListUsers(opt *gitlab.ListUsersOptions) ([]*gitlab.User, *gitlab.Response, error) {
	users, resp, err := u.service.ListUsers(opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return users, resp, nil
}

// NotesServiceWrapper wraps the real Notes service.
type NotesServiceWrapper struct {
	service gitlab.NotesServiceInterface
}

func (n *NotesServiceWrapper) CreateIssueNote(
	pid any,
	issue int64,
	opt *gitlab.CreateIssueNoteOptions,
) (*gitlab.Note, *gitlab.Response, error) {
	note, resp, err := n.service.CreateIssueNote(pid, issue, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return note, resp, nil
}

func (n *NotesServiceWrapper) CreateMergeRequestNote(
	pid any,
	mergeRequest int64,
	opt *gitlab.CreateMergeRequestNoteOptions,
) (*gitlab.Note, *gitlab.Response, error) {
	note, resp, err := n.service.CreateMergeRequestNote(pid, mergeRequest, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return note, resp, nil
}

// MergeRequestsServiceWrapper wraps the real MergeRequests service.
type MergeRequestsServiceWrapper struct {
	service gitlab.MergeRequestsServiceInterface
}

func (m *MergeRequestsServiceWrapper) CreateMergeRequest(
	pid any,
	opt *gitlab.CreateMergeRequestOptions,
) (*gitlab.MergeRequest, *gitlab.Response, error) {
	mr, resp, err := m.service.CreateMergeRequest(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return mr, resp, nil
}

// MilestonesServiceWrapper wraps the real Milestones service.
type MilestonesServiceWrapper struct {
	service gitlab.MilestonesServiceInterface
}

func (m *MilestonesServiceWrapper) ListMilestones(
	pid any,
	opt *gitlab.ListMilestonesOptions,
) ([]*gitlab.Milestone, *gitlab.Response, error) {
	milestones, resp, err := m.service.ListMilestones(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return milestones, resp, nil
}

// GroupsServiceWrapper wraps the real Groups service.
type GroupsServiceWrapper struct {
	service gitlab.GroupsServiceInterface
}

func (g *GroupsServiceWrapper) GetGroup(
	gid any,
	opt *gitlab.GetGroupOptions,
) (*gitlab.Group, *gitlab.Response, error) {
	group, resp, err := g.service.GetGroup(gid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return group, resp, nil
}

// EpicsServiceWrapper wraps the real Epics service.
type EpicsServiceWrapper struct {
	service gitlab.EpicsServiceInterface
}

func (e *EpicsServiceWrapper) ListGroupEpics(
	gid any,
	opt *gitlab.ListGroupEpicsOptions,
) ([]*gitlab.Epic, *gitlab.Response, error) {
	epics, resp, err := e.service.ListGroupEpics(gid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return epics, resp, nil
}

func (e *EpicsServiceWrapper) CreateEpic(
	gid any,
	opt *gitlab.CreateEpicOptions,
) (*gitlab.Epic, *gitlab.Response, error) {
	epic, resp, err := e.service.CreateEpic(gid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return epic, resp, nil
}
