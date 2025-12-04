package app

import (
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

func (m *MockGitLabClient) MergeRequests() MergeRequestsService {
	args := m.Called()
	result, _ := args.Get(0).(MergeRequestsService)
	return result
}

func (m *MockGitLabClient) Milestones() MilestonesService {
	args := m.Called()
	result, _ := args.Get(0).(MilestonesService)
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

// MockMergeRequestsService is a mock implementation of MergeRequestsService.
type MockMergeRequestsService struct {
	mock.Mock
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

// MockMilestonesService is a mock implementation of MilestonesService.
type MockMilestonesService struct {
	mock.Mock
}

func (m *MockMilestonesService) ListMilestones(
	pid any,
	opt *gitlab.ListMilestonesOptions,
) ([]*gitlab.Milestone, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	milestones, _ := args.Get(0).([]*gitlab.Milestone)
	response, _ := args.Get(1).(*gitlab.Response)
	return milestones, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}
