package app

import (
	"github.com/stretchr/testify/mock"
	"gitlab.com/gitlab-org/api/client-go"
)

// MockGitLabClient is a mock implementation of GitLabClient
type MockGitLabClient struct {
	mock.Mock
}

func (m *MockGitLabClient) Projects() ProjectsService {
	args := m.Called()
	return args.Get(0).(ProjectsService)
}

func (m *MockGitLabClient) Issues() IssuesService {
	args := m.Called()
	return args.Get(0).(IssuesService)
}

func (m *MockGitLabClient) Labels() LabelsService {
	args := m.Called()
	return args.Get(0).(LabelsService)
}

func (m *MockGitLabClient) Users() UsersService {
	args := m.Called()
	return args.Get(0).(UsersService)
}

// MockProjectsService is a mock implementation of ProjectsService
type MockProjectsService struct {
	mock.Mock
}

func (m *MockProjectsService) GetProject(pid interface{}, opt *gitlab.GetProjectOptions) (*gitlab.Project, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	return args.Get(0).(*gitlab.Project), args.Get(1).(*gitlab.Response), args.Error(2)
}

// MockIssuesService is a mock implementation of IssuesService
type MockIssuesService struct {
	mock.Mock
}

func (m *MockIssuesService) ListProjectIssues(pid interface{}, opt *gitlab.ListProjectIssuesOptions) ([]*gitlab.Issue, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	return args.Get(0).([]*gitlab.Issue), args.Get(1).(*gitlab.Response), args.Error(2)
}

func (m *MockIssuesService) CreateIssue(pid interface{}, opt *gitlab.CreateIssueOptions) (*gitlab.Issue, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	return args.Get(0).(*gitlab.Issue), args.Get(1).(*gitlab.Response), args.Error(2)
}

// MockLabelsService is a mock implementation of LabelsService
type MockLabelsService struct {
	mock.Mock
}

func (m *MockLabelsService) ListLabels(pid interface{}, opt *gitlab.ListLabelsOptions) ([]*gitlab.Label, *gitlab.Response, error) {
	args := m.Called(pid, opt)
	return args.Get(0).([]*gitlab.Label), args.Get(1).(*gitlab.Response), args.Error(2)
}

// MockUsersService is a mock implementation of UsersService
type MockUsersService struct {
	mock.Mock
}

func (m *MockUsersService) CurrentUser() (*gitlab.User, *gitlab.Response, error) {
	args := m.Called()
	return args.Get(0).(*gitlab.User), args.Get(1).(*gitlab.Response), args.Error(2)
}