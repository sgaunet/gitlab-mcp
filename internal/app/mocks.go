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

// MockProjectsService is a mock implementation of ProjectsService.
type MockProjectsService struct {
	mock.Mock
}

func (m *MockProjectsService) GetProject(
	pid interface{}, 
	opt *gitlab.GetProjectOptions,
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
	pid interface{}, 
	opt *gitlab.ListProjectIssuesOptions,
) ([]*gitlab.Issue, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	issues, _ := args.Get(0).([]*gitlab.Issue)
	response, _ := args.Get(1).(*gitlab.Response)
	return issues, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockIssuesService) CreateIssue(
	pid interface{}, 
	opt *gitlab.CreateIssueOptions,
) (*gitlab.Issue, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	issue, _ := args.Get(0).(*gitlab.Issue)
	response, _ := args.Get(1).(*gitlab.Response)
	return issue, response, args.Error(errorArgIndex) //nolint:wrapcheck // Mock should pass through errors
}

func (m *MockIssuesService) UpdateIssue(
	pid interface{}, 
	issue int, 
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
	pid interface{}, 
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