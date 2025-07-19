package app

import (
	"gitlab.com/gitlab-org/api/client-go"
)

// GitLabClientWrapper wraps the real GitLab client to implement our interfaces
type GitLabClientWrapper struct {
	client *gitlab.Client
}

// NewGitLabClient creates a new GitLab client wrapper
func NewGitLabClient(client *gitlab.Client) GitLabClient {
	return &GitLabClientWrapper{
		client: client,
	}
}

// Projects returns the Projects service
func (g *GitLabClientWrapper) Projects() ProjectsService {
	return &ProjectsServiceWrapper{service: g.client.Projects}
}

// Issues returns the Issues service
func (g *GitLabClientWrapper) Issues() IssuesService {
	return &IssuesServiceWrapper{service: g.client.Issues}
}

// Labels returns the Labels service
func (g *GitLabClientWrapper) Labels() LabelsService {
	return &LabelsServiceWrapper{service: g.client.Labels}
}

// Users returns the Users service
func (g *GitLabClientWrapper) Users() UsersService {
	return &UsersServiceWrapper{service: g.client.Users}
}

// ProjectsServiceWrapper wraps the real Projects service
type ProjectsServiceWrapper struct {
	service gitlab.ProjectsServiceInterface
}

func (p *ProjectsServiceWrapper) GetProject(pid interface{}, opt *gitlab.GetProjectOptions) (*gitlab.Project, *gitlab.Response, error) {
	return p.service.GetProject(pid, opt)
}

// IssuesServiceWrapper wraps the real Issues service
type IssuesServiceWrapper struct {
	service gitlab.IssuesServiceInterface
}

func (i *IssuesServiceWrapper) ListProjectIssues(pid interface{}, opt *gitlab.ListProjectIssuesOptions) ([]*gitlab.Issue, *gitlab.Response, error) {
	return i.service.ListProjectIssues(pid, opt)
}

func (i *IssuesServiceWrapper) CreateIssue(pid interface{}, opt *gitlab.CreateIssueOptions) (*gitlab.Issue, *gitlab.Response, error) {
	return i.service.CreateIssue(pid, opt)
}

func (i *IssuesServiceWrapper) UpdateIssue(pid interface{}, issue int, opt *gitlab.UpdateIssueOptions) (*gitlab.Issue, *gitlab.Response, error) {
	return i.service.UpdateIssue(pid, issue, opt)
}

// LabelsServiceWrapper wraps the real Labels service
type LabelsServiceWrapper struct {
	service gitlab.LabelsServiceInterface
}

func (l *LabelsServiceWrapper) ListLabels(pid interface{}, opt *gitlab.ListLabelsOptions) ([]*gitlab.Label, *gitlab.Response, error) {
	return l.service.ListLabels(pid, opt)
}

// UsersServiceWrapper wraps the real Users service
type UsersServiceWrapper struct {
	service gitlab.UsersServiceInterface
}

func (u *UsersServiceWrapper) CurrentUser() (*gitlab.User, *gitlab.Response, error) {
	return u.service.CurrentUser()
}