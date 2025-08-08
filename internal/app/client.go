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

// ProjectsServiceWrapper wraps the real Projects service.
type ProjectsServiceWrapper struct {
	service gitlab.ProjectsServiceInterface
}

func (p *ProjectsServiceWrapper) GetProject(
	pid interface{}, 
	opt *gitlab.GetProjectOptions,
) (*gitlab.Project, *gitlab.Response, error) {
	project, resp, err := p.service.GetProject(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return project, resp, nil
}

func (p *ProjectsServiceWrapper) EditProject(
	pid interface{}, 
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
	pid interface{}, 
	opt *gitlab.ListProjectIssuesOptions,
) ([]*gitlab.Issue, *gitlab.Response, error) {
	issues, resp, err := i.service.ListProjectIssues(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return issues, resp, nil
}

func (i *IssuesServiceWrapper) CreateIssue(
	pid interface{}, 
	opt *gitlab.CreateIssueOptions,
) (*gitlab.Issue, *gitlab.Response, error) {
	issue, resp, err := i.service.CreateIssue(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return issue, resp, nil
}

func (i *IssuesServiceWrapper) UpdateIssue(
	pid interface{}, 
	issue int, 
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
	pid interface{}, 
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
	pid interface{}, 
	issue int, 
	opt *gitlab.CreateIssueNoteOptions,
) (*gitlab.Note, *gitlab.Response, error) {
	note, resp, err := n.service.CreateIssueNote(pid, issue, opt)
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
	pid interface{}, 
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
	pid interface{}, 
	opt *gitlab.ListMilestonesOptions,
) ([]*gitlab.Milestone, *gitlab.Response, error) {
	milestones, resp, err := m.service.ListMilestones(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return milestones, resp, nil
}