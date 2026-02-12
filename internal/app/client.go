package app

import (
	"fmt"
	"io"

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

// Groups returns the Groups service.
func (g *GitLabClientWrapper) Groups() GroupsService {
	return &GroupsServiceWrapper{service: g.client.Groups}
}

// Epics returns the Epics service.
func (g *GitLabClientWrapper) Epics() EpicsService {
	return &EpicsServiceWrapper{service: g.client.Epics}
}

// EpicIssues returns the EpicIssues service.
func (g *GitLabClientWrapper) EpicIssues() EpicIssuesService {
	return &EpicIssuesServiceWrapper{service: g.client.EpicIssues}
}

// Pipelines returns the Pipelines service.
func (g *GitLabClientWrapper) Pipelines() PipelinesService {
	return &PipelinesServiceWrapper{service: g.client.Pipelines}
}

// Jobs returns the Jobs service.
func (g *GitLabClientWrapper) Jobs() JobsService {
	return &JobsServiceWrapper{service: g.client.Jobs}
}

// GroupLabels returns the GroupLabels service.
func (g *GitLabClientWrapper) GroupLabels() GroupLabelsService {
	return &GroupLabelsServiceWrapper{service: g.client.GroupLabels}
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

func (i *IssuesServiceWrapper) ListGroupIssues(
	gid any,
	opt *gitlab.ListGroupIssuesOptions,
	options ...gitlab.RequestOptionFunc,
) ([]*gitlab.Issue, *gitlab.Response, error) {
	issues, resp, err := i.service.ListGroupIssues(gid, opt, options...)
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

func (i *IssuesServiceWrapper) GetIssue(
	pid any,
	issue int,
) (*gitlab.Issue, *gitlab.Response, error) {
	iss, resp, err := i.service.GetIssue(pid, int64(issue), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return iss, resp, nil
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

// GroupLabelsServiceWrapper wraps the real GroupLabels service.
type GroupLabelsServiceWrapper struct {
	service gitlab.GroupLabelsServiceInterface
}

func (g *GroupLabelsServiceWrapper) ListGroupLabels(
	gid any,
	opt *gitlab.ListGroupLabelsOptions,
) ([]*gitlab.GroupLabel, *gitlab.Response, error) {
	labels, resp, err := g.service.ListGroupLabels(gid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return labels, resp, nil
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

// EpicIssuesServiceWrapper wraps the real EpicIssues service.
type EpicIssuesServiceWrapper struct {
	service gitlab.EpicIssuesServiceInterface
}

func (e *EpicIssuesServiceWrapper) AssignEpicIssue(
	gid any,
	epic, issue int64,
) (*gitlab.EpicIssueAssignment, *gitlab.Response, error) {
	epicIssue, resp, err := e.service.AssignEpicIssue(gid, epic, issue)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return epicIssue, resp, nil
}

// PipelinesServiceWrapper wraps the real Pipelines service.
type PipelinesServiceWrapper struct {
	service gitlab.PipelinesServiceInterface
}

func (p *PipelinesServiceWrapper) ListProjectPipelines(
	pid any,
	opt *gitlab.ListProjectPipelinesOptions,
) ([]*gitlab.PipelineInfo, *gitlab.Response, error) {
	pipelines, resp, err := p.service.ListProjectPipelines(pid, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return pipelines, resp, nil
}

// JobsServiceWrapper wraps the real Jobs service.
type JobsServiceWrapper struct {
	service gitlab.JobsServiceInterface
}

func (j *JobsServiceWrapper) ListPipelineJobs(
	pid any,
	pipelineID int64,
	opt *gitlab.ListJobsOptions,
) ([]*gitlab.Job, *gitlab.Response, error) {
	jobs, resp, err := j.service.ListPipelineJobs(pid, pipelineID, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return jobs, resp, nil
}

func (j *JobsServiceWrapper) GetTraceFile(
	pid any,
	jobID int64,
	options ...gitlab.RequestOptionFunc,
) (io.Reader, *gitlab.Response, error) {
	trace, resp, err := j.service.GetTraceFile(pid, jobID, options...)
	if err != nil {
		return nil, nil, fmt.Errorf("gitlab client: %w", err)
	}
	return trace, resp, nil
}
