package app

import (
	"io"

	"github.com/stretchr/testify/mock"
	"gitlab.com/gitlab-org/api/client-go"
)

const (
	errorArgIndex = 2 // Index for error argument in mock calls
)

// MockGitLabClient is a mock implementation of GitLabClient.
type MockGitLabClient struct {
	mock.Mock
}

func (m *MockGitLabClient) Projects() ProjectsService {
	args := m.Called()
	result, _ := args.Get(0).(ProjectsService)
	return result
}

func (m *MockGitLabClient) Issues() IssuesService {
	args := m.Called()
	result, _ := args.Get(0).(IssuesService)
	return result
}

func (m *MockGitLabClient) Labels() LabelsService {
	args := m.Called()
	result, _ := args.Get(0).(LabelsService)
	return result
}

func (m *MockGitLabClient) Users() UsersService {
	args := m.Called()
	result, _ := args.Get(0).(UsersService)
	return result
}

func (m *MockGitLabClient) Notes() NotesService {
	args := m.Called()
	result, _ := args.Get(0).(NotesService)
	return result
}

func (m *MockGitLabClient) Groups() GroupsService {
	args := m.Called()
	result, _ := args.Get(0).(GroupsService)
	return result
}

func (m *MockGitLabClient) Epics() EpicsService {
	args := m.Called()
	result, _ := args.Get(0).(EpicsService)
	return result
}

func (m *MockGitLabClient) EpicIssues() EpicIssuesService {
	args := m.Called()
	result, _ := args.Get(0).(EpicIssuesService)
	return result
}

func (m *MockGitLabClient) Pipelines() PipelinesService {
	args := m.Called()
	result, _ := args.Get(0).(PipelinesService)
	return result
}

func (m *MockGitLabClient) Jobs() JobsService {
	args := m.Called()
	result, _ := args.Get(0).(JobsService)
	return result
}

func (m *MockGitLabClient) GroupLabels() GroupLabelsService {
	args := m.Called()
	result, _ := args.Get(0).(GroupLabelsService)
	return result
}

func (m *MockGitLabClient) MergeRequests() MergeRequestsService {
	args := m.Called()
	result, _ := args.Get(0).(MergeRequestsService)
	return result
}

func (m *MockGitLabClient) MergeRequestApprovals() MergeRequestApprovalsService {
	args := m.Called()
	result, _ := args.Get(0).(MergeRequestApprovalsService)
	return result
}

// MockProjectsService is a mock implementation of ProjectsService.
type MockProjectsService struct {
	mock.Mock
}

func (m *MockProjectsService) GetProject(
	pid any,
	opt *gitlab.GetProjectOptions,
) (*gitlab.Project, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	project, _ := args.Get(0).(*gitlab.Project)
	response, _ := args.Get(1).(*gitlab.Response)
	return project, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockProjectsService) EditProject(
	pid any,
	opt *gitlab.EditProjectOptions,
) (*gitlab.Project, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	project, _ := args.Get(0).(*gitlab.Project)
	response, _ := args.Get(1).(*gitlab.Response)
	return project, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockIssuesService is a mock implementation of IssuesService.
type MockIssuesService struct {
	mock.Mock
}

func (m *MockIssuesService) ListProjectIssues(
	pid any,
	opt *gitlab.ListProjectIssuesOptions,
) ([]*gitlab.Issue, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	issues, _ := args.Get(0).([]*gitlab.Issue)
	response, _ := args.Get(1).(*gitlab.Response)
	return issues, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockIssuesService) ListGroupIssues(
	gid any,
	opt *gitlab.ListGroupIssuesOptions,
	options ...gitlab.RequestOptionFunc,
) ([]*gitlab.Issue, *gitlab.Response, error) {
	args := m.Called(gid, opt, options)
	issues, _ := args.Get(0).([]*gitlab.Issue)
	response, _ := args.Get(1).(*gitlab.Response)
	return issues, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockIssuesService) CreateIssue(
	pid any,
	opt *gitlab.CreateIssueOptions,
) (*gitlab.Issue, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	issue, _ := args.Get(0).(*gitlab.Issue)
	response, _ := args.Get(1).(*gitlab.Response)
	return issue, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockIssuesService) UpdateIssue(
	pid any,
	issue int64,
	opt *gitlab.UpdateIssueOptions,
) (*gitlab.Issue, *gitlab.Response, error) {
	args := m.Called(pid, issue, opt)
	updatedIssue, _ := args.Get(0).(*gitlab.Issue)
	response, _ := args.Get(1).(*gitlab.Response)
	return updatedIssue, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockIssuesService) GetIssue(
	pid any,
	issue int,
) (*gitlab.Issue, *gitlab.Response, error) {
	args := m.Called(pid, issue)
	iss, _ := args.Get(0).(*gitlab.Issue)
	response, _ := args.Get(1).(*gitlab.Response)
	return iss, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockLabelsService is a mock implementation of LabelsService.
type MockLabelsService struct {
	mock.Mock
}

func (m *MockLabelsService) ListLabels(
	pid any,
	opt *gitlab.ListLabelsOptions,
) ([]*gitlab.Label, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	labels, _ := args.Get(0).([]*gitlab.Label)
	response, _ := args.Get(1).(*gitlab.Response)
	return labels, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockGroupLabelsService is a mock implementation of GroupLabelsService.
type MockGroupLabelsService struct {
	mock.Mock
}

func (m *MockGroupLabelsService) ListGroupLabels(
	gid any,
	opt *gitlab.ListGroupLabelsOptions,
) ([]*gitlab.GroupLabel, *gitlab.Response, error) {
	args := m.Called(gid, opt)
	if args.Get(0) == nil {
		response, _ := args.Get(1).(*gitlab.Response)
		return nil, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
	}
	labels, _ := args.Get(0).([]*gitlab.GroupLabel)
	response, _ := args.Get(1).(*gitlab.Response)
	return labels, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockUsersService is a mock implementation of UsersService.
type MockUsersService struct {
	mock.Mock
}

func (m *MockUsersService) CurrentUser() (*gitlab.User, *gitlab.Response, error) {
	args := m.Called()
	user, _ := args.Get(0).(*gitlab.User)
	response, _ := args.Get(1).(*gitlab.Response)
	return user, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockUsersService) ListUsers(opt *gitlab.ListUsersOptions) ([]*gitlab.User, *gitlab.Response, error) {
	args := m.Called(opt)
	users, _ := args.Get(0).([]*gitlab.User)
	response, _ := args.Get(1).(*gitlab.Response)
	return users, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockNotesService is a mock implementation of NotesService.
type MockNotesService struct {
	mock.Mock
}

func (m *MockNotesService) CreateIssueNote(
	pid any,
	issue int64,
	opt *gitlab.CreateIssueNoteOptions,
) (*gitlab.Note, *gitlab.Response, error) {
	args := m.Called(pid, issue, opt)
	note, _ := args.Get(0).(*gitlab.Note)
	response, _ := args.Get(1).(*gitlab.Response)
	return note, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockNotesService) CreateMergeRequestNote(
	pid any,
	mergeRequest int64,
	opt *gitlab.CreateMergeRequestNoteOptions,
) (*gitlab.Note, *gitlab.Response, error) {
	args := m.Called(pid, mergeRequest, opt)
	note, _ := args.Get(0).(*gitlab.Note)
	response, _ := args.Get(1).(*gitlab.Response)
	return note, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockGroupsService is a mock implementation of GroupsService.
type MockGroupsService struct {
	mock.Mock
}

func (m *MockGroupsService) GetGroup(
	gid any,
	opt *gitlab.GetGroupOptions,
) (*gitlab.Group, *gitlab.Response, error) {
	args := m.Called(gid, opt)
	group, _ := args.Get(0).(*gitlab.Group)
	response, _ := args.Get(1).(*gitlab.Response)
	return group, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockEpicsService is a mock implementation of EpicsService.
type MockEpicsService struct {
	mock.Mock
}

func (m *MockEpicsService) ListGroupEpics(
	gid any,
	opt *gitlab.ListGroupEpicsOptions,
) ([]*gitlab.Epic, *gitlab.Response, error) {
	args := m.Called(gid, opt)
	epics, _ := args.Get(0).([]*gitlab.Epic)
	response, _ := args.Get(1).(*gitlab.Response)
	return epics, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockEpicsService) CreateEpic(
	gid any,
	opt *gitlab.CreateEpicOptions,
) (*gitlab.Epic, *gitlab.Response, error) {
	args := m.Called(gid, opt)
	if args.Get(0) == nil {
		response, _ := args.Get(1).(*gitlab.Response)
		return nil, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
	}
	epic, _ := args.Get(0).(*gitlab.Epic)
	response, _ := args.Get(1).(*gitlab.Response)
	return epic, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockEpicsService) UpdateEpic(
	gid any,
	epic int,
	opt *gitlab.UpdateEpicOptions,
) (*gitlab.Epic, *gitlab.Response, error) {
	args := m.Called(gid, epic, opt)
	if args.Get(0) == nil {
		response, _ := args.Get(1).(*gitlab.Response)
		return nil, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
	}
	updatedEpic, _ := args.Get(0).(*gitlab.Epic)
	response, _ := args.Get(1).(*gitlab.Response)
	return updatedEpic, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockEpicIssuesService is a mock implementation of EpicIssuesService.
type MockEpicIssuesService struct {
	mock.Mock
}

func (m *MockEpicIssuesService) AssignEpicIssue(
	gid any,
	epic, issue int64,
) (*gitlab.EpicIssueAssignment, *gitlab.Response, error) {
	args := m.Called(gid, epic, issue)
	if args.Get(0) == nil {
		response, _ := args.Get(1).(*gitlab.Response)
		return nil, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
	}
	epicIssue, _ := args.Get(0).(*gitlab.EpicIssueAssignment)
	response, _ := args.Get(1).(*gitlab.Response)
	return epicIssue, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockPipelinesService is a mock implementation of PipelinesService.
type MockPipelinesService struct {
	mock.Mock
}

func (m *MockPipelinesService) ListProjectPipelines(
	pid any,
	opt *gitlab.ListProjectPipelinesOptions,
) ([]*gitlab.PipelineInfo, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	pipelines, _ := args.Get(0).([]*gitlab.PipelineInfo)
	response, _ := args.Get(1).(*gitlab.Response)
	return pipelines, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockJobsService is a mock implementation of JobsService.
type MockJobsService struct {
	mock.Mock
}

func (m *MockJobsService) ListPipelineJobs(
	pid any,
	pipelineID int64,
	opt *gitlab.ListJobsOptions,
) ([]*gitlab.Job, *gitlab.Response, error) {
	args := m.Called(pid, pipelineID, opt)
	jobs, _ := args.Get(0).([]*gitlab.Job)
	response, _ := args.Get(1).(*gitlab.Response)
	return jobs, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockJobsService) GetTraceFile(
	pid any,
	jobID int64,
	options ...gitlab.RequestOptionFunc,
) (io.Reader, *gitlab.Response, error) {
	args := m.Called(pid, jobID, options)
	reader, _ := args.Get(0).(io.Reader)
	response, _ := args.Get(1).(*gitlab.Response)
	return reader, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockMergeRequestsService is a mock implementation of MergeRequestsService.
type MockMergeRequestsService struct {
	mock.Mock
}

func (m *MockMergeRequestsService) ListProjectMergeRequests(
	pid any,
	opt *gitlab.ListProjectMergeRequestsOptions,
) ([]*gitlab.MergeRequest, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	mrs, _ := args.Get(0).([]*gitlab.MergeRequest)
	response, _ := args.Get(1).(*gitlab.Response)
	return mrs, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockMergeRequestsService) CreateMergeRequest(
	pid any,
	opt *gitlab.CreateMergeRequestOptions,
) (*gitlab.MergeRequest, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	mr, _ := args.Get(0).(*gitlab.MergeRequest)
	response, _ := args.Get(1).(*gitlab.Response)
	return mr, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockMergeRequestsService) GetMergeRequest(
	pid any,
	mergeRequest int,
	opt *gitlab.GetMergeRequestsOptions,
) (*gitlab.MergeRequest, *gitlab.Response, error) {
	args := m.Called(pid, mergeRequest, opt)
	mr, _ := args.Get(0).(*gitlab.MergeRequest)
	response, _ := args.Get(1).(*gitlab.Response)
	return mr, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockMergeRequestsService) UpdateMergeRequest(
	pid any,
	mergeRequest int,
	opt *gitlab.UpdateMergeRequestOptions,
) (*gitlab.MergeRequest, *gitlab.Response, error) {
	args := m.Called(pid, mergeRequest, opt)
	mr, _ := args.Get(0).(*gitlab.MergeRequest)
	response, _ := args.Get(1).(*gitlab.Response)
	return mr, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockMergeRequestsService) AcceptMergeRequest(
	pid any,
	mergeRequest int,
	opt *gitlab.AcceptMergeRequestOptions,
) (*gitlab.MergeRequest, *gitlab.Response, error) {
	args := m.Called(pid, mergeRequest, opt)
	mr, _ := args.Get(0).(*gitlab.MergeRequest)
	response, _ := args.Get(1).(*gitlab.Response)
	return mr, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockMergeRequestsService) GetMergeRequestDiffVersions(
	pid any,
	mergeRequest int,
	opt *gitlab.GetMergeRequestDiffVersionsOptions,
) ([]*gitlab.MergeRequestDiffVersion, *gitlab.Response, error) {
	args := m.Called(pid, mergeRequest, opt)
	diffs, _ := args.Get(0).([]*gitlab.MergeRequestDiffVersion)
	response, _ := args.Get(1).(*gitlab.Response)
	return diffs, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockMergeRequestsService) ListMergeRequestDiffs(
	pid any,
	mergeRequest int,
	opt *gitlab.ListMergeRequestDiffsOptions,
) ([]*gitlab.MergeRequestDiff, *gitlab.Response, error) {
	args := m.Called(pid, mergeRequest, opt)
	diffs, _ := args.Get(0).([]*gitlab.MergeRequestDiff)
	response, _ := args.Get(1).(*gitlab.Response)
	return diffs, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

// MockMergeRequestApprovalsService is a mock implementation of MergeRequestApprovalsService.
type MockMergeRequestApprovalsService struct {
	mock.Mock
}

func (m *MockMergeRequestApprovalsService) ApproveMergeRequest(
	pid any,
	mergeRequest int64,
	opt *gitlab.ApproveMergeRequestOptions,
) (*gitlab.MergeRequestApprovals, *gitlab.Response, error) {
	args := m.Called(pid, mergeRequest, opt)
	approvals, _ := args.Get(0).(*gitlab.MergeRequestApprovals)
	response, _ := args.Get(1).(*gitlab.Response)
	return approvals, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}
